package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"SwiftRay/app/models"
)

type mockXray struct {
	mu         sync.Mutex
	installed  bool
	running    bool
	startErr   error
	validateErr error
	stopCalls  int
	startCalls int
}

func (m *mockXray) IsInstalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.installed
}

func (m *mockXray) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *mockXray) GetInstalledVersion() string { return "1.8.4" }

func (m *mockXray) ValidateResources() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.validateErr
}

func (m *mockXray) Start(onLogLine func(string, string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCalls++
	if m.startErr != nil {
		return m.startErr
	}
	m.running = true
	return nil
}

func (m *mockXray) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalls++
	m.running = false
}

type mockProxy struct {
	mu           sync.Mutex
	enabled      bool
	enableCalls  int
	disableCalls int
}

func (m *mockProxy) EnableProxy(httpPort, socksPort int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enableCalls++
	m.enabled = true
	return nil
}

func (m *mockProxy) DisableProxy() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disableCalls++
	m.enabled = false
	return nil
}

type mockSettings struct {
	mu       sync.Mutex
	settings models.AppSettings
}

func newMockSettings() *mockSettings {
	return &mockSettings{settings: models.DefaultSettings()}
}

func (m *mockSettings) LoadSettings() models.AppSettings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settings
}

func (m *mockSettings) SaveSettings(s models.AppSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	return nil
}

func TestConnectionManager_InitialState(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())
	if cm.GetState() != StateIdle {
		t.Fatalf("expected idle, got %s", cm.GetState())
	}
}

func TestConnectionManager_StateTransitions(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())

	states := []ConnectionState{StateConnecting, StateConnected, StateDisconnecting, StateIdle, StateError}
	for _, s := range states {
		cm.transition(StateConnecting, "test", "")
		cm.transition(s, "test", "")
		if cm.GetState() != s {
			t.Fatalf("expected %s, got %s", s, cm.GetState())
		}
	}
}

func TestConnectionManager_PreventDuplicateConnect(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())
	cm.transition(StateConnected, "already connected", "")

	err := cm.Connect("")
	if err == nil {
		t.Fatal("expected error on duplicate connect")
	}
}

func TestConnectionManager_PreventDuplicateDisconnect(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())

	err := cm.Disconnect()
	if err != nil {
		t.Fatalf("expected nil on disconnect when idle, got %v", err)
	}
}

func TestConnectionManager_ConcurrentConnect(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true, running: false}, &mockProxy{}, newMockSettings())

	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cm.Connect("")
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if errCount != 9 {
		t.Fatalf("expected 9 errors from concurrent connect, got %d", errCount)
	}
}

func TestConnectionManager_StateChangeCallback(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())

	var received []StateChange
	var mu sync.Mutex

	cm.OnStateChange(func(sc StateChange) {
		mu.Lock()
		received = append(received, sc)
		mu.Unlock()
	})

	cm.transition(StateConnecting, "step1", "")
	cm.transition(StateConnected, "step2", "")

	// Callbacks are async via goroutines; wait until we have at least 2
	deadline := time.After(2 * time.Second)
	for {
		mu.Lock()
		count := len(received)
		mu.Unlock()
		if count >= 2 {
			break
		}
		select {
		case <-deadline:
			mu.Lock()
			t.Fatalf("timed out waiting for callbacks, got %d", len(received))
			mu.Unlock()
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// Find the connecting and connected entries
	var foundConnecting, foundConnected bool
	for _, r := range received {
		if r.State == StateConnecting {
			foundConnecting = true
		}
		if r.State == StateConnected {
			foundConnected = true
		}
	}

	if !foundConnecting {
		t.Fatal("expected connecting callback")
	}
	if !foundConnected {
		t.Fatal("expected connected callback")
	}
}

func TestConnectionManager_ConnectValidatesResources(t *testing.T) {
	xray := &mockXray{installed: true, validateErr: fmt.Errorf("bundled xray not found")}
	proxy := &mockProxy{}
	store := newMockSettings()

	cm := NewConnectionManager(xray, proxy, store)

	err := cm.Connect("")
	if err == nil {
		t.Fatal("expected error when resources are invalid")
	}

	if cm.GetState() != StateError {
		t.Fatalf("expected error state, got %s", cm.GetState())
	}
}

func TestConnectionManager_FailedStartup(t *testing.T) {
	xray := &mockXray{installed: true, startErr: fmt.Errorf("binary corrupted")}
	proxy := &mockProxy{}
	store := newMockSettings()

	cm := NewConnectionManager(xray, proxy, store)

	err := cm.Connect("")
	if err == nil {
		t.Fatal("expected error on failed startup")
	}

	if cm.GetState() != StateError {
		t.Fatalf("expected error state, got %s", cm.GetState())
	}
}

func TestConnectionManager_DisconnectCleansUp(t *testing.T) {
	xray := &mockXray{installed: true, running: true}
	proxy := &mockProxy{enabled: true}
	store := newMockSettings()

	cm := NewConnectionManager(xray, proxy, store)
	cm.transition(StateConnected, "connected", "")

	err := cm.Disconnect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.GetState() != StateIdle {
		t.Fatalf("expected idle after disconnect, got %s", cm.GetState())
	}

	if xray.stopCalls == 0 {
		t.Fatal("expected xray.Stop to be called")
	}
}

func TestConnectionManager_RapidConnectDisconnect(t *testing.T) {
	xray := &mockXray{installed: true, running: true}
	proxy := &mockProxy{}
	store := newMockSettings()

	cm := NewConnectionManager(xray, proxy, store)
	cm.transition(StateConnected, "connected", "")

	err := cm.Disconnect()
	if err != nil {
		t.Fatalf("disconnect error: %v", err)
	}

	err = cm.Connect("")
	if err != nil {
		t.Fatalf("reconnect error: %v", err)
	}

	if cm.GetState() != StateConnected {
		t.Fatalf("expected connected, got %s", cm.GetState())
	}
}

func TestConnectionManager_LastError(t *testing.T) {
	cm := NewConnectionManager(&mockXray{installed: true}, &mockProxy{}, newMockSettings())

	cm.transition(StateError, "failed", "something went wrong")

	if cm.GetLastError() != "something went wrong" {
		t.Fatalf("expected error message, got %s", cm.GetLastError())
	}
}

func TestConnectionManager_RecoverStaleProcess(t *testing.T) {
	xray := &mockXray{installed: true, running: true}
	proxy := &mockProxy{}
	store := newMockSettings()

	cm := NewConnectionManager(xray, proxy, store)
	cm.state = StateConnected

	cm.RecoverStaleProcess()

	if cm.GetState() != StateIdle {
		t.Fatalf("expected idle after recovery, got %s", cm.GetState())
	}

	if xray.stopCalls == 0 {
		t.Fatal("expected stale process to be stopped")
	}
}
