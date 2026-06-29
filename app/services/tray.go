package services

import (
	_ "embed"
	"sync"

	"SwiftRay/internal/systray"
)

//go:embed tray_icon_disconnected.png
var iconDisconnected []byte

//go:embed tray_icon_connected.png
var iconConnected []byte

type TrayCallbacks struct {
	OnShow        func()
	OnToggleProxy func()
	OnConnect     func()
	OnDisconnect  func()
	OnRefreshIP   func()
	OnQuit        func()
}

type TrayManager struct {
	mu          sync.Mutex
	callbacks   TrayCallbacks
	isConnected bool
	ipAddress   string
	ipCountry   string
	proxyEnabled bool
	hasNode     bool
	canConnect  bool
	ready       bool
	started     bool

	showItem       *systray.MenuItem
	proxyItem      *systray.MenuItem
	connectItem    *systray.MenuItem
	disconnectItem *systray.MenuItem
	ipItem         *systray.MenuItem
	quitItem       *systray.MenuItem
}

func NewTrayManager() *TrayManager {
	return &TrayManager{}
}

func (t *TrayManager) SetCallbacks(cb TrayCallbacks) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callbacks = cb
}

func (t *TrayManager) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.mu.Unlock()

	go func() {
		systray.Run(t.onReady, t.onExit)
	}()
}

func (t *TrayManager) Stop() {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return
	}
	t.started = false
	t.ready = false
	t.mu.Unlock()

	systray.StopSystray()
}

func (t *TrayManager) ForceQuit() {
	systray.ForceQuit()
}

func (t *TrayManager) IsStarted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.started
}

func (t *TrayManager) onReady() {
	t.mu.Lock()
	cb := t.callbacks
	connected := t.isConnected
	ip := t.ipAddress
	country := t.ipCountry
	proxy := t.proxyEnabled
	t.mu.Unlock()

	if connected {
		systray.SetIcon(iconConnected)
	} else {
		systray.SetIcon(iconDisconnected)
	}
	systray.SetTitle("SwiftRay")
	systray.SetTooltip("SwiftRay - Proxy Client")

	t.showItem = systray.AddMenuItem("Show SwiftRay", "Bring window to front")
	systray.AddSeparator()
	t.proxyItem = systray.AddMenuItemCheckbox("System Proxy", "Toggle system proxy", proxy)
	systray.AddSeparator()
	t.connectItem = systray.AddMenuItem("Connect", "Connect to selected node")
	t.disconnectItem = systray.AddMenuItem("Disconnect", "Disconnect from proxy")
	systray.AddSeparator()

	ipLabel := t.formatIPLabel(ip, country)
	t.ipItem = systray.AddMenuItem(ipLabel, "IP Information")
	t.ipItem.Disable()

	systray.AddSeparator()
	t.quitItem = systray.AddMenuItem("Quit SwiftRay", "Quit the application")

	t.mu.Lock()
	t.ready = true
	t.mu.Unlock()

	t.updateMenuStates()

	go t.handleClicks(cb)
}

func (t *TrayManager) onExit() {}

func (t *TrayManager) formatIPLabel(ip, country string) string {
	if ip == "" {
		return "IP: --"
	}
	if country != "" {
		return ip + " (" + country + ")"
	}
	return ip
}

func (t *TrayManager) updateMenuStates() {
	t.mu.Lock()
	ready := t.ready
	connected := t.isConnected
	proxy := t.proxyEnabled
	node := t.hasNode
	canConn := t.canConnect
	ip := t.ipAddress
	country := t.ipCountry
	t.mu.Unlock()

	if !ready {
		return
	}

	if connected {
		t.connectItem.Hide()
		t.disconnectItem.Show()
	} else {
		t.connectItem.Show()
		t.disconnectItem.Hide()
		if !node || !canConn {
			t.connectItem.Disable()
		} else {
			t.connectItem.Enable()
		}
	}

	if proxy {
		t.proxyItem.Check()
	} else {
		t.proxyItem.Uncheck()
	}

	t.ipItem.SetTitle(t.formatIPLabel(ip, country))
}

func (t *TrayManager) handleClicks(cb TrayCallbacks) {
	for {
		select {
		case <-t.showItem.ClickedCh:
			if cb.OnShow != nil {
				cb.OnShow()
			}
		case <-t.proxyItem.ClickedCh:
			if cb.OnToggleProxy != nil {
				cb.OnToggleProxy()
			}
		case <-t.connectItem.ClickedCh:
			if cb.OnConnect != nil {
				cb.OnConnect()
			}
		case <-t.disconnectItem.ClickedCh:
			if cb.OnDisconnect != nil {
				cb.OnDisconnect()
			}
		case <-t.ipItem.ClickedCh:
			if cb.OnRefreshIP != nil {
				cb.OnRefreshIP()
			}
		case <-t.quitItem.ClickedCh:
			if cb.OnQuit != nil {
				cb.OnQuit()
			}
			return
		}
	}
}

func (t *TrayManager) SetConnected(connected bool) {
	t.mu.Lock()
	t.isConnected = connected
	t.mu.Unlock()
}

func (t *TrayManager) SetIPAddress(ip string) {
	t.mu.Lock()
	t.ipAddress = ip
	t.mu.Unlock()
}

func (t *TrayManager) SetIPCountry(country string) {
	t.mu.Lock()
	t.ipCountry = country
	t.mu.Unlock()
}

func (t *TrayManager) SetProxyEnabled(enabled bool) {
	t.mu.Lock()
	t.proxyEnabled = enabled
	t.mu.Unlock()
}

func (t *TrayManager) SetHasNode(has bool) {
	t.mu.Lock()
	t.hasNode = has
	t.mu.Unlock()
}

func (t *TrayManager) SetCanConnect(can bool) {
	t.mu.Lock()
	t.canConnect = can
	t.mu.Unlock()
}

func (t *TrayManager) NotifyStateChanged() {
	t.mu.Lock()
	ready := t.ready
	connected := t.isConnected
	t.mu.Unlock()

	if !ready {
		return
	}

	if connected {
		systray.SetIcon(iconConnected)
	} else {
		systray.SetIcon(iconDisconnected)
	}
	t.updateMenuStates()
}
