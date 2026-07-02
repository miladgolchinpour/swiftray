package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"SwiftRay/app/models"
	"SwiftRay/app/platform"
	"SwiftRay/app/services"
	"SwiftRay/app/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx          context.Context
	storage      *services.StorageService
	xray         *services.XrayService
	subscription *services.SubscriptionService
	conn         *services.ConnectionManager
	logStream    *services.LogStream
	ipInfo       *services.IPInfoService
	updater      *services.UpdaterService
	tray         *services.TrayManager

	startupComplete bool
	quitting        bool
}

type APIResponse struct {
	OK      bool        `json:"ok"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type AppInfo struct {
	StartupComplete bool   `json:"startupComplete"`
	XrayInstalled   bool   `json:"xrayInstalled"`
	XrayVersion     string `json:"xrayVersion"`
	GeoIPReady      bool   `json:"geoIPReady"`
	GeoSiteReady    bool   `json:"geoSiteReady"`
	DataDir         string `json:"dataDir"`
}

type ConnectionStateDTO struct {
	State   string `json:"state"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type URLTestResult struct {
	NodeID string `json:"nodeId"`
	Delay  *int   `json:"delay"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("[SwiftRay] Starting up...")

	a.storage = services.NewStorageService()
	a.xray = services.NewXrayService(a.storage)
	a.subscription = services.NewSubscriptionService()
	a.logStream = services.NewLogStream(200)
	a.ipInfo = services.NewIPInfoService()
	a.updater = services.NewUpdaterService()
	a.tray = services.NewTrayManager()

	// Wire updater progress to Wails events
	a.updater.OnProgress(func(p services.DownloadProgress) {
		if a.ctx != nil {
			data, _ := json.Marshal(p)
			runtime.EventsEmit(a.ctx, "updater:progress", string(data))
		}
	})

	// Wire log streaming to Wails events
	a.logStream.OnLog(func(entry models.LogEntry) {
		if a.ctx != nil {
			data, _ := json.Marshal(entry)
			runtime.EventsEmit(a.ctx, "log:entry", string(data))
		}
	})

	// Wire IP info updates to Wails events
	a.ipInfo.OnUpdate(func(info models.IPInfo) {
		if a.ctx != nil {
			data, _ := json.Marshal(info)
			runtime.EventsEmit(a.ctx, "ipinfo:update", string(data))
		}
	})

	// Wire connection manager to emit state changes and log entries
	pm := platform.NewProxyManager()
	a.conn = services.NewConnectionManager(a.xray, pm, a.storage)
	a.conn.OnStateChange(func(sc services.StateChange) {
		if a.ctx != nil {
			data, _ := json.Marshal(sc)
			runtime.EventsEmit(a.ctx, "connection:state-change", string(data))
		}
		// Log state changes
		if sc.Error != "" {
			a.logStream.AddFromXray("error", fmt.Sprintf("Connection error: %s - %s", sc.Message, sc.Error))
		} else {
			a.logStream.AddFromXray("info", sc.Message)
		}
		// Update tray state immediately
		a.updateTrayState()
		// Schedule delayed IP refresh after connect/disconnect (matches Swift behavior)
		if sc.State == services.StateConnected || sc.State == services.StateIdle {
			go func() {
				time.Sleep(2 * time.Second)
				a.fetchAndEmitIPInfo()
				a.updateTrayState()
			}()
		}
	})

	// Wire xray process output to log stream
	a.xray.SetLogCallback(func(level, message string) {
		a.logStream.AddFromXray(level, message)
	})

	if err := a.xray.EnsureDirectory(); err != nil {
		log.Printf("[SwiftRay] Failed to create app directory: %v", err)
	}

	a.conn.RecoverStaleProcess()
	a.startupComplete = true
	log.Println("[SwiftRay] Startup complete")

	// Set up system tray callbacks (always set, Start is conditional)
	a.setupTrayCallbacks()

	// Start tray only if enableMenuBar is true
	settings := a.storage.LoadSettings()
	if settings.EnableMenuBar {
		a.tray.Start()
		a.updateTrayState()
	}
}

func (a *App) updateTrayState() {
	if a.tray == nil || !a.tray.IsStarted() {
		return
	}
	connected := a.conn.GetState() == services.StateConnected
	a.tray.SetConnected(connected)

	settings := a.storage.LoadSettings()
	a.tray.SetProxyEnabled(settings.EnableSystemProxy)

	selectedID := a.storage.LoadSelectedNodeID()
	a.tray.SetHasNode(selectedID != "")
	a.tray.SetCanConnect(a.xray.ValidateResources() == nil)

	info := a.ipInfo.GetCached()
	if info != nil && info.IPAddress != "" {
		a.tray.SetIPAddress(info.IPAddress)
		a.tray.SetIPCountry(info.CountryName)
	} else {
		a.tray.SetIPAddress("")
		a.tray.SetIPCountry("")
	}

	a.tray.NotifyStateChanged()
}

func (a *App) setupTrayCallbacks() {
	a.tray.SetCallbacks(services.TrayCallbacks{
		OnShow: func() {
			runtime.WindowShow(a.ctx)
		},
		OnToggleProxy: func() {
			settings := a.storage.LoadSettings()
			settings.EnableSystemProxy = !settings.EnableSystemProxy
			a.storage.SaveSettings(settings)
			a.tray.SetProxyEnabled(settings.EnableSystemProxy)
			a.tray.NotifyStateChanged()
			if a.ctx != nil {
				data, _ := json.Marshal(map[string]interface{}{
					"enableSystemProxy": settings.EnableSystemProxy,
				})
				runtime.EventsEmit(a.ctx, "settings:proxy-toggled", string(data))
			}
			go func() {
				pm := platform.NewProxyManager()
				if settings.EnableSystemProxy && a.conn.GetState() == services.StateConnected {
					httpPort := settings.SOCKSPort
					if !settings.MixedPort {
						httpPort = settings.HTTPPort
					}
					pm.EnableProxy(httpPort, settings.SOCKSPort)
				} else {
					pm.DisableProxy()
				}
			}()
		},
		OnConnect: func() {
			selectedID := a.storage.LoadSelectedNodeID()
			if selectedID == "" {
				a.logStream.AddFromXray("error", "No node selected")
				return
			}
			a.Connect()
		},
		OnDisconnect: func() {
			a.Disconnect()
		},
		OnRefreshIP: func() {
			go a.fetchAndEmitIPInfo()
		},
		OnQuit: func() {
			a.requestQuit()
		},
	})
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("[SwiftRay] Shutting down...")
	a.conn.Disconnect()
	// Don't call a.tray.Stop() here — the status item is automatically
	// removed when the app terminates. Calling stop_systray during shutdown
	// crashes because the NSApplication run loop is already stopped.
}

func (a *App) requestQuit() {
	a.quitting = true
	a.conn.Disconnect()
	settings := a.storage.LoadSettings()
	if settings.EnableSystemProxy {
		platform.NewProxyManager().DisableProxy()
	}
	if a.tray != nil {
		a.tray.Stop()
	}
	runtime.Quit(a.ctx)
}

func okResponse(data interface{}) APIResponse {
	return APIResponse{OK: true, Data: data}
}

func okMessage(msg string) APIResponse {
	return APIResponse{OK: true, Message: msg}
}

func errResponse(err string) APIResponse {
	return APIResponse{OK: false, Error: err}
}

// --- System ---

func (a *App) GetAppInfo() APIResponse {
	status := a.xray.GetResourceStatus()
	info := AppInfo{
		StartupComplete: a.startupComplete,
		XrayInstalled:   status.XrayExists,
		XrayVersion:     a.xray.GetInstalledVersion(),
		GeoIPReady:      status.GeoIPExists,
		GeoSiteReady:    status.GeoSiteExists,
		DataDir:         utils.AppDataDir(),
	}
	return okResponse(info)
}

func (a *App) VerifyReady() APIResponse {
	if !a.startupComplete {
		return errResponse("application not started")
	}
	if err := a.xray.ValidateResources(); err != nil {
		return errResponse(fmt.Sprintf("runtime resources invalid: %v", err))
	}
	return okMessage("ready")
}

func (a *App) GetResourceStatus() APIResponse {
	return okResponse(a.xray.GetResourceStatus())
}

// --- Settings ---

func (a *App) GetSettings() APIResponse {
	settings := a.storage.LoadSettings()
	return okResponse(settings)
}

func (a *App) SaveSettings(settings models.AppSettings) APIResponse {
	oldSettings := a.storage.LoadSettings()
	if err := a.storage.SaveSettings(settings); err != nil {
		log.Printf("[SwiftRay] Save settings error: %v", err)
		return errResponse(fmt.Sprintf("failed to save settings: %v", err))
	}
	log.Println("[SwiftRay] Settings saved")

	// Handle system proxy toggle — apply/unapply immediately
	if oldSettings.EnableSystemProxy != settings.EnableSystemProxy {
		go func() {
			pm := platform.NewProxyManager()
			if settings.EnableSystemProxy && a.conn.GetState() == services.StateConnected {
				httpPort := settings.SOCKSPort
				if !settings.MixedPort {
					httpPort = settings.HTTPPort
				}
				pm.EnableProxy(httpPort, settings.SOCKSPort)
			} else {
				pm.DisableProxy()
			}
		}()
		// Sync tray state
		if a.tray != nil && a.tray.IsStarted() {
			a.tray.SetProxyEnabled(settings.EnableSystemProxy)
			a.tray.NotifyStateChanged()
		}
	}

	// Handle menu bar toggle
	if oldSettings.EnableMenuBar != settings.EnableMenuBar {
		if settings.EnableMenuBar {
			a.setupTrayCallbacks()
			a.tray.Start()
			a.updateTrayState()
		} else {
			a.tray.Stop()
		}
	}

	return okMessage("settings saved")
}

func (a *App) SaveAndReload(settings models.AppSettings) APIResponse {
	oldSettings := a.storage.LoadSettings()

	// Step 1: Save settings
	if err := a.storage.SaveSettings(settings); err != nil {
		log.Printf("[SwiftRay] Save settings error: %v", err)
		return errResponse(fmt.Sprintf("failed to save settings: %v", err))
	}
	log.Println("[SwiftRay] Settings saved")

	// Handle system proxy toggle — apply/unapply immediately
	if oldSettings.EnableSystemProxy != settings.EnableSystemProxy {
		go func() {
			pm := platform.NewProxyManager()
			if settings.EnableSystemProxy && a.conn.GetState() == services.StateConnected {
				httpPort := settings.SOCKSPort
				if !settings.MixedPort {
					httpPort = settings.HTTPPort
				}
				pm.EnableProxy(httpPort, settings.SOCKSPort)
			} else {
				pm.DisableProxy()
			}
		}()
		if a.tray != nil && a.tray.IsStarted() {
			a.tray.SetProxyEnabled(settings.EnableSystemProxy)
			a.tray.NotifyStateChanged()
		}
	}

	// Handle menu bar toggle
	if oldSettings.EnableMenuBar != settings.EnableMenuBar {
		if settings.EnableMenuBar {
			a.setupTrayCallbacks()
			a.tray.Start()
			a.updateTrayState()
		} else {
			a.tray.Stop()
		}
	}

	// Step 2: If not connected, just save
	if a.conn.GetState() != services.StateConnected {
		return okMessage("settings saved")
	}

	// Step 3: Regenerate config with new settings
	selectedID := a.storage.LoadSelectedNodeID()
	if selectedID == "" {
		return okMessage("settings saved (no node selected for reload)")
	}

	var selectedNode *models.Node
	subID := a.storage.LoadSelectedSubID()
	if subID != "" {
		subs := a.storage.LoadSubscriptions()
		for _, sub := range subs {
			if sub.ID == subID {
				for i, n := range sub.Nodes {
					if n.ID == selectedID {
						selectedNode = &sub.Nodes[i]
						break
					}
				}
				break
			}
		}
	} else {
		nodes := a.storage.LoadLocalNodes()
		for i, n := range nodes {
			if n.ID == selectedID {
				selectedNode = &nodes[i]
				break
			}
		}
	}

	if selectedNode == nil {
		return okMessage("settings saved (selected node not found for reload)")
	}

	// Step 4: Stop current connection
	if err := a.conn.Disconnect(); err != nil {
		log.Printf("[SwiftRay] Reload disconnect warning: %v", err)
	}

	// Step 5: Regenerate and write config
 freshSettings := a.storage.LoadSettings()
	config := services.GenerateConfig(*selectedNode, freshSettings)
	if err := a.xray.WriteConfig(config); err != nil {
		return okMessage(fmt.Sprintf("settings saved but config write failed: %v", err))
	}

	// Step 6: Reconnect
	if err := a.conn.Connect(utils.ConfigPath()); err != nil {
		return okMessage(fmt.Sprintf("settings saved but reconnect failed: %v", err))
	}

	log.Println("[SwiftRay] Settings saved and reloaded")
	go func() {
		time.Sleep(1 * time.Second)
		a.fetchAndEmitIPInfo()
	}()

	return okMessage("settings applied")
}

// --- Local Nodes ---

func (a *App) GetLocalNodes() APIResponse {
	nodes := a.storage.LoadLocalNodes()
	return okResponse(nodes)
}

func (a *App) SaveLocalNodes(nodes []models.Node) APIResponse {
	if err := a.storage.SaveLocalNodes(nodes); err != nil {
		log.Printf("[SwiftRay] Save local nodes error: %v", err)
		return errResponse(fmt.Sprintf("failed to save nodes: %v", err))
	}
	return okMessage("nodes saved")
}

func (a *App) AddLocalNode(node models.Node) APIResponse {
	nodes := a.storage.LoadLocalNodes()
	nodes = append(nodes, node)
	if err := a.storage.SaveLocalNodes(nodes); err != nil {
		return errResponse(fmt.Sprintf("failed to save node: %v", err))
	}
	return okResponse(nodes)
}

func (a *App) AddLocalNodes(nodes []models.Node) APIResponse {
	existing := a.storage.LoadLocalNodes()
	existing = append(existing, nodes...)
	if err := a.storage.SaveLocalNodes(existing); err != nil {
		return errResponse(fmt.Sprintf("failed to save nodes: %v", err))
	}
	return okResponse(existing)
}

func (a *App) UpdateLocalNode(node models.Node) APIResponse {
	nodes := a.storage.LoadLocalNodes()
	for i, n := range nodes {
		if n.ID == node.ID {
			nodes[i] = node
			break
		}
	}
	if err := a.storage.SaveLocalNodes(nodes); err != nil {
		return errResponse(fmt.Sprintf("failed to save node: %v", err))
	}
	return okResponse(nodes)
}

func (a *App) DeleteLocalNode(id string) APIResponse {
	nodes := a.storage.LoadLocalNodes()
	filtered := make([]models.Node, 0, len(nodes))
	for _, n := range nodes {
		if n.ID != id {
			filtered = append(filtered, n)
		}
	}
	if err := a.storage.SaveLocalNodes(filtered); err != nil {
		return errResponse(fmt.Sprintf("failed to delete node: %v", err))
	}
	return okResponse(filtered)
}

func (a *App) GetSelectedNodeID() APIResponse {
	id := a.storage.LoadSelectedNodeID()
	return okResponse(id)
}

func (a *App) SetSelectedNodeID(id string) APIResponse {
	if err := a.storage.SaveSelectedNodeID(id); err != nil {
		return errResponse(fmt.Sprintf("failed to save selection: %v", err))
	}
	return okMessage("selection saved")
}

// --- Subscriptions ---

func (a *App) GetSubscriptions() APIResponse {
	subs := a.storage.LoadSubscriptions()
	return okResponse(subs)
}

func (a *App) AddSubscription(name string, url string) APIResponse {
	subs := a.storage.LoadSubscriptions()
	sub := models.NewSubscription(name, url)

	nodes, err := a.subscription.Fetch(url)
	if err != nil {
		log.Printf("[SwiftRay] Subscription fetch warning: %v (saving subscription anyway)", err)
		sub.LastUpdated = nil
	} else {
		sub.Nodes = nodes
		now := time.Now().Format(time.RFC3339)
		sub.LastUpdated = &now
	}

	subs = append(subs, sub)
	if err := a.storage.SaveSubscriptions(subs); err != nil {
		return errResponse(fmt.Sprintf("failed to save subscription: %v", err))
	}
	log.Printf("[SwiftRay] Added subscription '%s' with %d nodes", name, len(sub.Nodes))
	return okResponse(subs)
}

func (a *App) UpdateSubscription(id string, name string, url string) APIResponse {
	subs := a.storage.LoadSubscriptions()
	for i, s := range subs {
		if s.ID == id {
			subs[i].Name = name
			subs[i].URL = url

			nodes, err := a.subscription.Fetch(url)
			if err != nil {
				log.Printf("[SwiftRay] Subscription refetch warning: %v", err)
			} else {
				subs[i].Nodes = nodes
				now := time.Now().Format(time.RFC3339)
				subs[i].LastUpdated = &now
			}
			break
		}
	}
	if err := a.storage.SaveSubscriptions(subs); err != nil {
		return errResponse(fmt.Sprintf("failed to save subscription: %v", err))
	}
	return okResponse(subs)
}

func (a *App) DeleteSubscription(id string) APIResponse {
	subs := a.storage.LoadSubscriptions()
	filtered := make([]models.Subscription, 0, len(subs))
	for _, s := range subs {
		if s.ID != id {
			filtered = append(filtered, s)
		}
	}
	if err := a.storage.SaveSubscriptions(filtered); err != nil {
		return errResponse(fmt.Sprintf("failed to delete subscription: %v", err))
	}
	return okResponse(filtered)
}

func (a *App) RefreshSubscription(id string) APIResponse {
	subs := a.storage.LoadSubscriptions()
	found := false
	for i, s := range subs {
		if s.ID == id {
			nodes, err := a.subscription.Fetch(s.URL)
			if err != nil {
				log.Printf("[SwiftRay] Refresh subscription error: %v", err)
				return errResponse(fmt.Sprintf("refresh failed: %v", err))
			}
			subs[i].Nodes = nodes
			now := time.Now().Format(time.RFC3339)
			subs[i].LastUpdated = &now
			found = true
			log.Printf("[SwiftRay] Refreshed subscription '%s': %d nodes", s.Name, len(nodes))
			break
		}
	}
	if !found {
		return errResponse("subscription not found")
	}
	if err := a.storage.SaveSubscriptions(subs); err != nil {
		return errResponse(fmt.Sprintf("failed to save: %v", err))
	}
	return okResponse(subs)
}

func (a *App) RefreshAllSubscriptions() APIResponse {
	subs := a.storage.LoadSubscriptions()
	updated := 0
	for i, s := range subs {
		nodes, err := a.subscription.Fetch(s.URL)
		if err != nil {
			log.Printf("[SwiftRay] Refresh all warning for '%s': %v", s.Name, err)
			continue
		}
		subs[i].Nodes = nodes
		now := time.Now().Format(time.RFC3339)
		subs[i].LastUpdated = &now
		updated++
	}
	if err := a.storage.SaveSubscriptions(subs); err != nil {
		return errResponse(fmt.Sprintf("failed to save: %v", err))
	}
	log.Printf("[SwiftRay] Refreshed %d/%d subscriptions", updated, len(subs))
	return okResponse(subs)
}

func (a *App) GetSelectedSubID() APIResponse {
	id := a.storage.LoadSelectedSubID()
	return okResponse(id)
}

func (a *App) SetSelectedSubID(id string) APIResponse {
	if err := a.storage.SaveSelectedSubID(id); err != nil {
		return errResponse(fmt.Sprintf("failed to save selection: %v", err))
	}
	return okMessage("selection saved")
}

// --- Connection ---

func (a *App) GetConnectionState() APIResponse {
	state := a.conn.GetState()
	dto := ConnectionStateDTO{
		State: string(state),
	}
	if state == services.StateError {
		dto.Error = a.conn.GetLastError()
	}
	return okResponse(dto)
}

func (a *App) Connect() APIResponse {
	selectedID := a.storage.LoadSelectedNodeID()
	if selectedID == "" {
		return errResponse("no node selected")
	}

	var selectedNode *models.Node

	subID := a.storage.LoadSelectedSubID()
	if subID != "" {
		subs := a.storage.LoadSubscriptions()
		for _, sub := range subs {
			if sub.ID == subID {
				for i, n := range sub.Nodes {
					if n.ID == selectedID {
						selectedNode = &sub.Nodes[i]
						break
					}
				}
				break
			}
		}
	} else {
		nodes := a.storage.LoadLocalNodes()
		for i, n := range nodes {
			if n.ID == selectedID {
				selectedNode = &nodes[i]
				break
			}
		}
	}

	if selectedNode == nil {
		return errResponse("selected node not found")
	}

	settings := a.storage.LoadSettings()
	config := services.GenerateConfig(*selectedNode, settings)
	if err := a.xray.WriteConfig(config); err != nil {
		return errResponse(fmt.Sprintf("failed to write config: %v", err))
	}

	err := a.conn.Connect(utils.ConfigPath())
	if err != nil {
		return errResponse(err.Error())
	}

	// Fetch IP info after connection
	go func() {
		time.Sleep(1 * time.Second)
		a.fetchAndEmitIPInfo()
	}()

	return okMessage("connected")
}

func (a *App) Disconnect() APIResponse {
	err := a.conn.Disconnect()
	if err != nil {
		return errResponse(err.Error())
	}

	// Fetch IP info after disconnection
	go func() {
		time.Sleep(1 * time.Second)
		a.fetchAndEmitIPInfo()
	}()

	return okMessage("disconnected")
}

func (a *App) fetchAndEmitIPInfo() {
	settings := a.storage.LoadSettings()
	socksPort := settings.SOCKSPort

	isConnected := a.conn.GetState() == services.StateConnected
	info, err := a.ipInfo.FetchForConnection(isConnected, socksPort)
	if err != nil {
		a.logStream.AddFromXray("warning", fmt.Sprintf("Failed to fetch IP info: %v", err))
		return
	}

	if a.ctx != nil {
		data, _ := json.Marshal(info)
		runtime.EventsEmit(a.ctx, "ipinfo:update", string(data))
	}

	// Update tray with new IP info
	a.updateTrayState()
}

// --- Updates ---

func (a *App) GetXrayVersion() APIResponse {
	version := a.xray.GetInstalledVersion()
	return okResponse(version)
}

func (a *App) CheckXrayUpdate() APIResponse {
	available, err := a.xray.CheckForUpdate()
	if err != nil {
		return errResponse(fmt.Sprintf("check failed: %v", err))
	}
	return okResponse(available)
}

// --- Runtime Updater ---

func (a *App) CheckXrayUpdateStatus() APIResponse {
	return okResponse(a.updater.CheckXrayUpdate())
}

func (a *App) DownloadXrayUpdate() APIResponse {
	if a.updater.IsDownloading() {
		return errResponse("download already in progress")
	}

	go func() {
		err := a.updater.DownloadXray(nil)
		if err != nil {
			log.Printf("[Updater] Runtime download failed: %v", err)
			if a.ctx != nil {
				data, _ := json.Marshal(services.DownloadProgress{Type: "runtime", Stage: "error", Status: "error", Error: err.Error()})
				runtime.EventsEmit(a.ctx, "updater:progress", string(data))
			}
		}
	}()

	return okMessage("download started")
}

func (a *App) CancelDownload() APIResponse {
	a.updater.Cancel()
	return okMessage("download cancelled")
}

// --- Logs ---

func (a *App) GetLogs() APIResponse {
	return okResponse(a.logStream.GetEntries())
}

func (a *App) ClearLogs() APIResponse {
	a.logStream.Clear()
	return okMessage("logs cleared")
}

// --- IP Info ---

func (a *App) GetIPInfo() APIResponse {
	info := a.ipInfo.GetCached()
	return okResponse(info)
}

func (a *App) FetchIPInfo() APIResponse {
	info, err := a.ipInfo.FetchDirect()
	if err != nil {
		return errResponse(fmt.Sprintf("fetch failed: %v", err))
	}
	return okResponse(info)
}

// --- URL Test ---

func (a *App) URLTest() APIResponse {
	settings := a.storage.LoadSettings()

	var nodes []models.Node
	subID := a.storage.LoadSelectedSubID()
	if subID != "" {
		subs := a.storage.LoadSubscriptions()
		for _, sub := range subs {
			if sub.ID == subID {
				nodes = sub.Nodes
				break
			}
		}
	} else {
		nodes = a.storage.LoadLocalNodes()
	}

	if len(nodes) == 0 {
		return errResponse("no nodes to test")
	}

	return a.runURLTest(nodes, settings)
}

func (a *App) URLTestLocal() APIResponse {
	settings := a.storage.LoadSettings()
	nodes := a.storage.LoadLocalNodes()

	if len(nodes) == 0 {
		return errResponse("no local nodes to test")
	}

	return a.runURLTest(nodes, settings)
}

func (a *App) URLTestNodes(nodes []models.Node) APIResponse {
	if len(nodes) == 0 {
		return errResponse("no nodes to test")
	}

	settings := a.storage.LoadSettings()
	return a.runURLTest(nodes, settings)
}

func (a *App) runURLTest(nodes []models.Node, settings models.AppSettings) APIResponse {
	a.logStream.AddFromXray("info", fmt.Sprintf("Starting URL test for %d nodes...", len(nodes)))

	go func() {
		a.xray.URLTest(nodes, settings.URLTestTimeout, settings.URLTestConcurrency, settings.URLTestMode, func(nodeID string, delay *int) {
			result := URLTestResult{NodeID: nodeID, Delay: delay}
			data, _ := json.Marshal(result)
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "urltest:result", string(data))
			}
		})
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "urltest:complete", "")
		}
		a.logStream.AddFromXray("info", "URL test complete")
	}()

	return okMessage("url test started")
}


