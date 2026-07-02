import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { api } from '@/api/client'
import { useNotifications } from '@/composables/useNotifications'
import { useConnectionStore } from './connection'
import type { Node, Subscription, AppSettings, IPInfo, LogEntry } from '@/types'

export const useAppStore = defineStore('app', () => {
  const notify = useNotifications()

  const selectedNodeID = ref<string | null>(null)
  const selectedSubID = ref<string | null>(null)
  const logs = ref<LogEntry[]>([])
  const ipInfo = ref<IPInfo | null>(null)

  const settings = ref<AppSettings>({
    httpPort: 2080,
    socksPort: 2080,
    mixedPort: true,
    enableUDP: true,
    allowLAN: true,
    routeOnly: false,
    enableSniffing: true,
    sniffHTTP: true,
    sniffTLS: true,
    sniffQUIC: true,
    sniffFakeDNS: false,
    useProxyAuth: false,
    proxyUsername: '',
    proxyPassword: '',
    defaultFingerprint: 'chrome',
    enableFragment: false,
    fragmentPackLength: '100-200',
    fragmentSleep: '50-100',
    fragmentInterval: '1-5',
    localDNS: '8.8.8.8',
    remoteDNS: 'https://cloudflare-dns.com/dns-query',
    bootstrapDNS: '8.8.8.8',
    parallelQuery: true,
    serveStale: true,
    useSystemHosts: true,
    customDNSHosts: false,
    fakeIP: false,
    blockSVCBHTTPS: false,
    validateRegionalDomain: 0,
    enableSystemProxy: true,
    enableMenuBar: true,
    routingMode: 0,
    domainStrategy: 1,
    bypassIran: false,
    bypassRussia: false,
    bypassChina: false,
    pingTestURL: 'https://www.google.com/generate_204',
    customGeoSources: '',
    exclusions: 'localhost\n127.0.0.0/8\n::1',
    urlTestMode: 'tcp',
    urlTestTimeout: 3,
    urlTestConcurrency: 8,
  })

  const subscriptions = ref<Subscription[]>([])
  const localNodes = ref<Node[]>([])
  const refreshingSubID = ref<string | null>(null)
  const refreshingAll = ref(false)
  const isSaving = ref(false)
  const isReloading = ref(false)

  // --- Debounced auto-save with live reload ---
  let saveTimer: ReturnType<typeof setTimeout> | null = null
  let loadComplete = false

  function scheduleSave() {
    if (!loadComplete) return
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => {
      performSaveAndReload()
    }, 1800)
  }

  async function performSaveAndReload() {
    isSaving.value = true
    try {
      const connectionStore = useConnectionStore()
      const wasConnected = connectionStore.isConnected

      if (wasConnected) {
        isReloading.value = true
        const { error } = await api.saveAndReload(settings.value as any)
        isReloading.value = false
        if (error) {
          notify.error('Settings reload failed', error)
        } else {
          notify.success('Settings applied')
        }
      } else {
        const { error } = await api.saveSettings(settings.value as any)
        if (error) {
          notify.error('Save failed', error)
        } else {
          notify.success('Settings saved')
        }
      }
    } finally {
      isSaving.value = false
    }
  }

  watch(settings, () => {
    scheduleSave()
  }, { deep: true })

  async function loadAll() {
    // Load settings first so other operations have correct values
    await loadSettings()
    // Load nodes and subscriptions first so validation has data to check against
    await Promise.all([
      loadLocalNodes(),
      loadSubscriptions(),
      loadSelectedSubID(),
    ])
    // Now load selected node ID — validateSelection needs nodes/subscriptions loaded
    await loadSelectedNodeID()
    // Mark load complete so auto-save can start
    loadComplete = true
  }

  async function loadSettings() {
    const { data, error } = await api.getSettings()
    if (data) {
      // Merge with defaults to ensure all fields are present
      const defaults = { ...settings.value }
      const loaded = data as unknown as Partial<AppSettings>
      settings.value = { ...defaults, ...loaded }
    } else if (error) {
      notify.handleAPIError(error, 'Load settings')
    }
  }

  async function loadLocalNodes() {
    const { data, error } = await api.getLocalNodes()
    if (data) {
      localNodes.value = data as unknown as Node[]
      validateSelection()
    } else if (error) {
      notify.handleAPIError(error, 'Load nodes')
    }
  }

  async function addLocalNode(node: Node) {
    const { data, error } = await api.addLocalNode(node as any)
    if (data) {
      localNodes.value = data as unknown as Node[]
      notify.success('Node added', node.name)
    } else if (error) {
      notify.handleAPIError(error, 'Add node')
    }
  }

  async function addLocalNodes(nodes: Node[]) {
    const { data, error } = await api.addLocalNodes(nodes as any)
    if (data) {
      localNodes.value = data as unknown as Node[]
    } else if (error) {
      notify.handleAPIError(error, 'Add nodes')
    }
  }

  async function updateLocalNode(node: Node) {
    const { data, error } = await api.updateLocalNode(node as any)
    if (data) {
      localNodes.value = data as unknown as Node[]
    } else if (error) {
      notify.handleAPIError(error, 'Update node')
    }
  }

  async function deleteLocalNode(id: string) {
    const { data, error } = await api.deleteLocalNode(id)
    if (data) {
      localNodes.value = data as unknown as Node[]
      if (selectedNodeID.value === id) {
        selectedNodeID.value = null
      }
    } else if (error) {
      notify.handleAPIError(error, 'Delete node')
    }
  }

  async function loadSelectedNodeID() {
    const { data } = await api.getSelectedNodeID()
    // Normalize: empty string or falsy → null
    selectedNodeID.value = (data && data !== '') ? data : null
    validateSelection()
  }

  function validateSelection() {
    if (!selectedNodeID.value) return

    if (selectedSubID.value) {
      const sub = subscriptions.value.find(s => s.id === selectedSubID.value)
      if (sub && !sub.nodes.some(n => n.id === selectedNodeID.value)) {
        selectedNodeID.value = null
      }
    } else {
      if (!localNodes.value.some(n => n.id === selectedNodeID.value)) {
        selectedNodeID.value = null
      }
    }
  }

  async function setSelectedNodeID(id: string) {
    // Normalize empty string to null for clean state
    selectedNodeID.value = id || null
    await api.setSelectedNodeID(id)
  }

  async function loadSubscriptions() {
    const { data, error } = await api.getSubscriptions()
    if (data) {
      const oldSub = selectedSubID.value ? subscriptions.value.find(s => s.id === selectedSubID.value) : null
      const oldSelectedNode = getSelectedNode()
      subscriptions.value = data as unknown as Subscription[]
      // Re-match selection after data load
      if (oldSub) {
        const newSub = subscriptions.value.find(s => s.name === oldSub.name)
        if (newSub) selectedSubID.value = newSub.id
      }
      if (selectedNodeID.value && oldSelectedNode && selectedSubID.value) {
        const sub = subscriptions.value.find(s => s.id === selectedSubID.value)
        if (sub) {
          const matched = sub.nodes.find(n => n.address === oldSelectedNode.address && n.port === oldSelectedNode.port)
          if (matched) selectedNodeID.value = matched.id
        }
      }
    } else if (error) {
      notify.handleAPIError(error, 'Load subscriptions')
    }
  }

  async function addSubscription(name: string, url: string) {
    const { data, error } = await api.addSubscription(name, url)
    if (data) {
      subscriptions.value = data as unknown as Subscription[]
      notify.success('Subscription added', name)
    } else if (error) {
      notify.handleAPIError(error, 'Add subscription')
    }
  }

  async function updateSubscription(id: string, name: string, url: string) {
    const { data, error } = await api.updateSubscription(id, name, url)
    if (data) {
      subscriptions.value = data as unknown as Subscription[]
    } else if (error) {
      notify.handleAPIError(error, 'Update subscription')
    }
  }

  async function deleteSubscription(id: string) {
    const { data, error } = await api.deleteSubscription(id)
    if (data) {
      subscriptions.value = data as unknown as Subscription[]
      if (selectedSubID.value === id) {
        selectedSubID.value = null
      }
    } else if (error) {
      notify.handleAPIError(error, 'Delete subscription')
    }
  }

  async function refreshSubscription(id: string) {
    refreshingSubID.value = id
    const { data, error } = await api.refreshSubscription(id)
    if (data) {
      const oldSub = subscriptions.value.find(s => s.id === id)
      const oldSelectedNode = getSelectedNode()
      subscriptions.value = data as unknown as Subscription[]
      reMatchSelectionAfterRefresh(oldSub, oldSelectedNode)
      notify.success('Subscription refreshed')
    } else if (error) {
      notify.handleAPIError(error, 'Refresh subscription')
    }
    refreshingSubID.value = null
  }

  async function refreshAllSubscriptions() {
    refreshingAll.value = true
    const { data, error } = await api.refreshAllSubscriptions()
    if (data) {
      const oldSub = selectedSubID.value ? subscriptions.value.find(s => s.id === selectedSubID.value) : null
      const oldSelectedNode = getSelectedNode()
      subscriptions.value = data as unknown as Subscription[]
      reMatchSelectionAfterRefresh(oldSub, oldSelectedNode)
      notify.success('All subscriptions refreshed')
    } else if (error) {
      notify.handleAPIError(error, 'Refresh all subscriptions')
    }
    refreshingAll.value = false
  }

  function getSelectedNode(): { address: string; port: number } | null {
    if (!selectedNodeID.value) return null
    if (selectedSubID.value) {
      const sub = subscriptions.value.find(s => s.id === selectedSubID.value)
      const node = sub?.nodes.find(n => n.id === selectedNodeID.value)
      return node ? { address: node.address, port: node.port } : null
    }
    const node = localNodes.value.find(n => n.id === selectedNodeID.value)
    return node ? { address: node.address, port: node.port } : null
  }

  function reMatchSelectionAfterRefresh(
    oldSub: Subscription | undefined | null,
    oldNode: { address: string; port: number } | null
  ) {
    // Re-match subscription by name (names are stable across refresh)
    if (selectedSubID.value && oldSub) {
      const newSub = subscriptions.value.find(s => s.name === oldSub.name)
      if (newSub) {
        selectedSubID.value = newSub.id
      } else {
        selectedSubID.value = null
      }
    }

    // Re-match node by address+port (stable across refresh)
    if (selectedNodeID.value && oldNode && selectedSubID.value) {
      const sub = subscriptions.value.find(s => s.id === selectedSubID.value)
      if (sub) {
        const matched = sub.nodes.find(n => n.address === oldNode.address && n.port === oldNode.port)
        if (matched) {
          selectedNodeID.value = matched.id
        } else {
          selectedNodeID.value = null
        }
      }
    }
  }

  async function loadSelectedSubID() {
    const { data } = await api.getSelectedSubID()
    if (data) {
      selectedSubID.value = data || null
    }
  }

  async function setSelectedSubID(id: string) {
    selectedSubID.value = id
    await api.setSelectedSubID(id)
  }

  function addLog(level: LogEntry['level'], message: string) {
    const entry: LogEntry = {
      id: crypto.randomUUID(),
      level,
      message,
      timestamp: new Date().toISOString(),
    }
    logs.value.unshift(entry)
    if (logs.value.length > 200) {
      logs.value = logs.value.slice(0, 200)
    }
  }

  function clearLogs() {
    logs.value = []
  }

  return {
    selectedNodeID,
    selectedSubID,
    logs,
    ipInfo,
    settings,
    subscriptions,
    localNodes,
    refreshingSubID,
    refreshingAll,
    isSaving,
    isReloading,
    loadAll,
    loadSettings,
    loadLocalNodes,
    addLocalNode,
    addLocalNodes,
    updateLocalNode,
    deleteLocalNode,
    setSelectedNodeID,
    validateSelection,
    loadSubscriptions,
    addSubscription,
    updateSubscription,
    deleteSubscription,
    refreshSubscription,
    refreshAllSubscriptions,
    setSelectedSubID,
    addLog,
    clearLogs,
  }
})
