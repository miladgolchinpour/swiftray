package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"SwiftRay/app/models"
)

type ConnectionState string

const (
	StateIdle          ConnectionState = "idle"
	StateConnecting    ConnectionState = "connecting"
	StateConnected     ConnectionState = "connected"
	StateDisconnecting ConnectionState = "disconnecting"
	StateError         ConnectionState = "error"
)

type StateChange struct {
	State   ConnectionState `json:"state"`
	Message string          `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type XrayProcess interface {
	IsInstalled() bool
	IsRunning() bool
	GetInstalledVersion() string
	ValidateResources() error
	Start(onLogLine func(string, string)) error
	Stop()
}

type ProxyControl interface {
	EnableProxy(httpPort, socksPort int) error
	DisableProxy() error
}

type SettingsStore interface {
	LoadSettings() models.AppSettings
	SaveSettings(models.AppSettings) error
}

type ConnectionManager struct {
	mu            sync.Mutex
	state         ConnectionState
	xray          XrayProcess
	proxy         ProxyControl
	storage       SettingsStore
	lastError     string
	onStateChange func(StateChange)
}

func NewConnectionManager(xray XrayProcess, proxy ProxyControl, storage SettingsStore) *ConnectionManager {
	return &ConnectionManager{
		state:   StateIdle,
		xray:    xray,
		proxy:   proxy,
		storage: storage,
	}
}

func (cm *ConnectionManager) OnStateChange(fn func(StateChange)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onStateChange = fn
}

func (cm *ConnectionManager) GetState() ConnectionState {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.state
}

func (cm *ConnectionManager) GetLastError() string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.lastError
}

func (cm *ConnectionManager) transition(state ConnectionState, msg string, errStr string) {
	cm.state = state
	if errStr != "" {
		cm.lastError = errStr
	}
	sc := StateChange{State: state, Message: msg, Error: errStr}
	if cm.onStateChange != nil {
		go cm.onStateChange(sc)
	}
	log.Printf("[Connection] → %s: %s", state, msg)
}

func (cm *ConnectionManager) Connect(configPath string) error {
	cm.mu.Lock()
	if cm.state == StateConnecting || cm.state == StateConnected || cm.state == StateDisconnecting {
		cm.mu.Unlock()
		return fmt.Errorf("cannot connect: current state is %s", cm.state)
	}
	cm.transition(StateConnecting, "Starting connection...", "")
	cm.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			cm.mu.Lock()
			cm.transition(StateError, fmt.Sprintf("panic: %v", r), fmt.Sprintf("%v", r))
			cm.mu.Unlock()
		}
	}()

	// Step 1: Validate bundled resources
	cm.mu.Lock()
	cm.transition(StateConnecting, "Validating resources...", "")
	cm.mu.Unlock()
	if err := cm.xray.ValidateResources(); err != nil {
		cm.mu.Lock()
		cm.transition(StateError, "Bundled resources missing", err.Error())
		cm.mu.Unlock()
		return fmt.Errorf("validate resources: %w", err)
	}

	// Step 2: Start Xray
	cm.mu.Lock()
	cm.transition(StateConnecting, "Starting Xray process...", "")
	cm.mu.Unlock()

	err := cm.xray.Start(func(level, message string) {})
	if err != nil {
		cm.mu.Lock()
		cm.transition(StateError, "Failed to start Xray", err.Error())
		cm.mu.Unlock()
		return fmt.Errorf("start xray: %w", err)
	}

	// Step 3: Verify process started
	time.Sleep(200 * time.Millisecond)
	if !cm.xray.IsRunning() {
		cm.mu.Lock()
		cm.transition(StateError, "Xray process exited immediately", "process not running after start")
		cm.mu.Unlock()
		return fmt.Errorf("xray process exited immediately")
	}

	// Step 4: Enable system proxy
	settings := cm.storage.LoadSettings()
	if settings.EnableSystemProxy {
		cm.mu.Lock()
		cm.transition(StateConnecting, "Configuring system proxy...", "")
		cm.mu.Unlock()

		// Match Swift: mixedPort means HTTP port = SOCKS port
		httpPort := settings.SOCKSPort
		socksPort := settings.SOCKSPort
		if !settings.MixedPort {
			httpPort = settings.HTTPPort
		}
		if err := cm.proxy.EnableProxy(httpPort, socksPort); err != nil {
			cm.mu.Lock()
			cm.transition(StateError, "Failed to configure system proxy", err.Error())
			cm.mu.Unlock()
			cm.xray.Stop()
			return fmt.Errorf("enable proxy: %w", err)
		}
	}

	cm.mu.Lock()
	cm.transition(StateConnected, "Connected", "")
	cm.mu.Unlock()

	return nil
}

func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	if cm.state == StateIdle || cm.state == StateDisconnecting {
		cm.mu.Unlock()
		return nil
	}
	if cm.state == StateConnecting {
		cm.mu.Unlock()
		return fmt.Errorf("cannot disconnect while connecting")
	}
	cm.transition(StateDisconnecting, "Disconnecting...", "")
	cm.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			cm.mu.Lock()
			cm.transition(StateError, fmt.Sprintf("panic during disconnect: %v", r), fmt.Sprintf("%v", r))
			cm.mu.Unlock()
		}
	}()

	settings := cm.storage.LoadSettings()
	if settings.EnableSystemProxy {
		if err := cm.proxy.DisableProxy(); err != nil {
			log.Printf("[Connection] Warning: failed to disable proxy: %v", err)
		}
	}

	cm.xray.Stop()

	cm.mu.Lock()
	cm.transition(StateIdle, "Disconnected", "")
	cm.mu.Unlock()

	return nil
}

func (cm *ConnectionManager) RecoverStaleProcess() {
	if cm.xray.IsRunning() {
		log.Println("[Connection] Found stale Xray process, cleaning up...")
		cm.xray.Stop()
		time.Sleep(200 * time.Millisecond)
	}

	cm.mu.Lock()
	if cm.state != StateIdle {
		cm.transition(StateIdle, "Recovered from stale state", "")
	}
	cm.mu.Unlock()
}
