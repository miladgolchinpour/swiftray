<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <!-- Header -->
        <div class="space-y-1">
          <h1 class="text-[28px] font-bold tracking-tight text-foreground">Dashboard</h1>
          <p class="text-sm text-muted-foreground">Monitor and control your connection</p>
        </div>

        <!-- Install Banner -->
        <div
          v-if="!resourceReady"
          class="flex items-center gap-3 p-4 rounded-xl bg-gradient-to-br from-red-500/90 to-red-500/70"
        >
          <AlertTriangle class="w-5 h-5 text-white flex-shrink-0" />
          <div class="flex-1">
            <p class="text-sm font-semibold text-white">Xray core or Geo files are not installed</p>
            <p class="text-xs text-white/80">Go to Settings → Xray to check bundled resources</p>
          </div>
          <Settings class="w-4 h-4 text-white/70" />
        </div>

        <!-- Unsaved Changes Banner -->
        <div
          v-if="connectionStore.isConnected && hasUnsavedChanges"
          class="flex items-center gap-3 p-4 rounded-xl bg-orange-500/8 border border-orange-500/30"
        >
          <AlertCircle class="w-5 h-5 text-orange-500 flex-shrink-0" />
          <span class="text-sm font-medium text-foreground">You have unsaved settings changes</span>
          <span class="flex-1" />
          <button
            v-if="isReloading"
            class="inline-flex items-center gap-1.5 text-sm font-medium text-blue-500"
          >
            <Loader2 class="w-4 h-4 animate-spin" />
            Reloading...
          </button>
          <button
            v-else
            class="h-8 px-4 rounded-lg bg-blue-600 text-xs font-semibold text-white hover:bg-blue-700 transition-colors duration-150"
            @click="saveAndReload"
          >
            Save & Reload
          </button>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-2 gap-4">
          <!-- Connection Card -->
          <div class="rounded-2xl border border-border bg-card p-[18px] shadow-sm space-y-3.5">
            <div class="flex items-center justify-between">
              <div
                class="w-10 h-10 rounded-[10px] flex items-center justify-center transition-all duration-500"
                :class="connectionStore.isConnected
                  ? 'bg-gradient-to-br from-green-500/80 to-green-500/50 shadow-md shadow-green-500/20'
                  : 'bg-gradient-to-br from-red-500/80 to-red-500/50'"
              >
                <Zap class="w-5 h-5 text-white" :stroke-width="2" />
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs font-medium text-muted-foreground">System</span>
                <button
                  class="relative w-9 h-[20px] rounded-full transition-colors duration-150"
                  :class="appStore.settings.enableSystemProxy ? 'bg-blue-600' : 'bg-muted'"
                  role="switch"
                  :aria-checked="appStore.settings.enableSystemProxy"
                  @click="toggleSystemProxy"
                >
                  <span
                    class="absolute top-[2px] left-[2px] w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-150"
                    :class="{ 'translate-x-[16px]': appStore.settings.enableSystemProxy }"
                  />
                </button>
              </div>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">Connection</p>
              <p class="text-lg font-bold text-foreground">{{ statusLabel }}</p>
            </div>
          </div>

          <!-- Current Node Card -->
          <div class="rounded-2xl border border-border bg-card p-[18px] shadow-sm space-y-3.5">
            <div
              class="w-10 h-10 rounded-[10px] flex items-center justify-center bg-gradient-to-br from-blue-500/80 to-blue-500/50"
            >
              <Server class="w-5 h-5 text-white" :stroke-width="2" />
            </div>
            <div>
              <p class="text-xs text-muted-foreground">Current Node</p>
              <p class="text-lg font-bold text-foreground truncate">
                {{ selectedNode?.name || 'None' }}
              </p>
            </div>
          </div>
        </div>

        <!-- IP Info Card -->
        <div class="rounded-2xl border border-border bg-card p-4 shadow-sm flex items-center gap-3.5">
          <div class="w-9 h-9 rounded-[9px] flex items-center justify-center bg-gradient-to-br from-indigo-500 to-indigo-500/50 flex-shrink-0">
            <Globe class="w-[18px] h-[18px] text-white" :stroke-width="2" />
          </div>
          <div class="flex-1 min-w-0">
            <template v-if="ipInfoStore.loading && !ipInfoStore.info">
              <div class="flex items-center gap-1.5">
                <Loader2 class="w-3.5 h-3.5 animate-spin text-muted-foreground" />
                <span class="text-sm text-muted-foreground">Fetching IP info...</span>
              </div>
            </template>
            <template v-else-if="ipInfoStore.info">
              <div class="flex items-center gap-1.5">
                <span class="text-sm">{{ countryFlag(ipInfoStore.info.countryCode) }}</span>
                <span class="text-sm font-semibold text-foreground">{{ ipInfoStore.info.ipAddress || '--' }}</span>
                <template v-if="ipInfoStore.info.cityName && ipInfoStore.info.countryName">
                  <span class="text-muted-foreground">·</span>
                  <span class="text-sm text-muted-foreground">{{ ipInfoStore.info.cityName }}, {{ ipInfoStore.info.countryName }}</span>
                </template>
              </div>
              <p v-if="ipInfoStore.info.asnOrganization" class="text-xs text-muted-foreground/70 truncate">
                {{ ipInfoStore.info.asnOrganization }}
              </p>
            </template>
            <template v-else>
              <p class="text-sm font-semibold text-foreground">--</p>
              <p class="text-xs text-muted-foreground/70">Tap refresh to fetch IP info</p>
            </template>
          </div>
          <button
            class="w-7 h-7 rounded-lg border border-border bg-card flex items-center justify-center text-muted-foreground hover:bg-accent transition-colors duration-150"
            title="IP details"
            @click="showIPDetail = true"
          >
            <Info class="w-3.5 h-3.5" />
          </button>
          <button
            class="w-7 h-7 rounded-lg border border-border bg-card flex items-center justify-center text-muted-foreground hover:bg-accent transition-colors duration-150"
            title="Refresh IP info"
            :disabled="ipInfoStore.loading"
            @click="ipInfoStore.fetch()"
          >
            <Loader2 v-if="ipInfoStore.loading" class="w-3.5 h-3.5 animate-spin" />
            <RefreshCw v-else class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- Action Buttons -->
        <div class="space-y-2">
          <div class="grid grid-cols-4 gap-3">
            <button
              :disabled="(!resourceReady && !connectionStore.isConnected) || connectionStore.isBusy || (!connectionStore.isConnected && !appStore.selectedNodeID)"
              class="h-10 rounded-lg text-sm font-semibold transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              :class="connectionStore.isConnected
                ? 'bg-red-600 text-white hover:bg-red-700'
                : 'bg-green-600 text-white hover:bg-green-700'"
              @click="handleConnect"
            >
              {{ connectionStore.isConnected ? 'Disconnect' : 'Connect' }}
            </button>
            <button
              :disabled="!connectionStore.isConnected || isReloading"
              class="h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-1.5"
            >
              <Loader2 v-if="isReloading" class="w-4 h-4 animate-spin" />
              <RefreshCw v-else class="w-4 h-4" />
              Reload
            </button>
            <button
              :disabled="urlTestRunning"
              class="h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-1.5"
              @click="handleURLTest"
            >
              <Radio class="w-4 h-4" :class="{ 'animate-pulse': urlTestRunning }" />
              {{ urlTestRunning ? 'Testing...' : 'URL Test' }}
            </button>
            <button
              :disabled="!appStore.selectedSubID"
              class="h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-1.5"
              @click="handleUpdateSub"
            >
              <RefreshCw class="w-4 h-4" />
              Update Sub
            </button>
          </div>
          <p
            v-if="!connectionStore.isConnected && !connectionStore.isConnecting && !appStore.selectedNodeID && resourceReady"
            class="text-xs text-red-500 flex items-center gap-1.5"
          >
            <AlertTriangle class="w-3 h-3" />
            Please select a node before connecting
          </p>
        </div>

        <!-- URL Test Progress -->
        <UrlTestProgress
          v-if="urlTestRunning"
          :tested="urlTestTested"
          :total="urlTestTotal"
          @cancel="cancelURLTest"
        />

        <!-- Node Section -->
        <div class="rounded-2xl border border-border bg-card shadow-sm overflow-hidden">
          <!-- Header -->
          <div class="flex items-center gap-2.5 px-[18px] py-[18px]">
            <div class="w-7 h-7 rounded-[7px] flex items-center justify-center flex-shrink-0 bg-gradient-to-br from-purple-500 to-purple-500/60">
              <Server class="w-3.5 h-3.5 text-white" :stroke-width="2" />
            </div>
            <h3 class="text-sm font-semibold text-card-foreground">Select Node</h3>
            <span class="flex-1" />
            <button
              class="h-7 px-3 rounded-md border border-border bg-card text-xs font-medium text-card-foreground hover:bg-accent transition-colors duration-150 inline-flex items-center gap-1.5"
              @click="showSourcePicker = true"
            >
              {{ sourceLabel }}
              <ChevronRight class="w-3 h-3" />
            </button>
          </div>
          <div class="border-t border-border" />

          <!-- Empty State -->
          <div v-if="sortedNodes.length === 0" class="flex flex-col items-center py-8 gap-3 bg-muted/20">
            <Server class="w-10 h-10 text-muted-foreground/30" />
            <p class="text-sm font-semibold text-muted-foreground">No Nodes Available</p>
            <p class="text-xs text-muted-foreground/70">Add a profile and select a node to get started</p>
          </div>

          <!-- Node List -->
          <div v-else>
            <div
              v-for="node in visibleNodes"
              :key="node.id"
              class="flex items-center gap-2.5 px-3.5 py-2.5 cursor-pointer select-none transition-colors duration-150 hover:bg-accent/50"
              :class="{ 'bg-accent/30': appStore.selectedNodeID === node.id }"
              @click="appStore.setSelectedNodeID(node.id)"
            >
              <div class="flex-1 min-w-0">
                <p class="text-[13px] font-medium text-card-foreground truncate">{{ node.name }}</p>
                <p class="text-xs text-muted-foreground truncate">
                  {{ protocolDisplayName(node.protocolType) }} - {{ node.address }} - {{ node.port }}
                </p>
              </div>
              <CheckCircle v-if="appStore.selectedNodeID === node.id" class="w-4 h-4 text-green-500 flex-shrink-0" />
              <span class="text-xs font-mono flex-shrink-0" :class="delayColorClass(node.delay)">
                {{ delayString(node.delay) }}
              </span>
              <button
                class="w-7 h-7 rounded-lg border border-border bg-card flex items-center justify-center text-muted-foreground hover:bg-accent transition-colors duration-150 flex-shrink-0"
                title="Export config"
                @click.stop="openExport(node)"
              >
                <Share class="w-3 h-3" />
              </button>
            </div>

            <!-- Show All/Less -->
            <template v-if="sortedNodes.length > maxVisibleNodes">
              <div class="border-t border-border" />
              <button
                class="w-full py-2.5 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors duration-150"
                @click="showAllNodes = !showAllNodes"
              >
                {{ showAllNodes ? 'Show less' : `Show all ${sortedNodes.length} nodes` }}
              </button>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- Source Picker Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showSourcePicker"
          class="fixed inset-0 z-50 flex items-center justify-center"
          @keydown.esc="showSourcePicker = false"
        >
          <div class="fixed inset-0 bg-black/50 backdrop-blur-sm" @click="showSourcePicker = false" />
          <div class="relative z-10 bg-card border border-border rounded-2xl shadow-lg flex flex-col overflow-hidden w-[420px]">
            <div class="flex flex-col items-center pt-7 pb-5 px-6">
              <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3 bg-foreground/10">
                <User class="w-7 h-7 text-foreground" :stroke-width="1.5" />
              </div>
              <h2 class="text-lg font-bold text-card-foreground">Select Source</h2>
            </div>
            <div class="border-t border-border" />

            <div class="flex-1 overflow-y-auto">
              <!-- Empty State -->
              <div v-if="appStore.localNodes.length === 0 && appStore.subscriptions.length === 0" class="flex flex-col items-center py-8 gap-3">
                <Inbox class="w-8 h-8 text-muted-foreground/30" />
                <p class="text-sm font-medium text-muted-foreground">No sources available</p>
                <p class="text-xs text-muted-foreground/70">Add subscriptions or local nodes first</p>
              </div>

              <template v-else>
                <!-- Local Nodes -->
                <button
                  v-if="appStore.localNodes.length > 0"
                  class="w-full flex items-center gap-3 px-5 py-3 hover:bg-accent/50 transition-colors duration-150 text-left"
                  @click="selectSource(null)"
                >
                  <Server class="w-4 h-4 text-purple-500 flex-shrink-0" />
                  <div class="flex-1">
                    <p class="text-sm font-medium text-foreground">Local Nodes</p>
                    <p class="text-xs text-muted-foreground">{{ appStore.localNodes.length }} nodes</p>
                  </div>
                  <CheckCircle v-if="!appStore.selectedSubID" class="w-4 h-4 text-green-500" />
                </button>
                <div v-if="appStore.localNodes.length > 0" class="border-t border-border ml-10" />

                <!-- Subscriptions -->
                <template v-for="sub in appStore.subscriptions" :key="sub.id">
                  <button
                    class="w-full flex items-center gap-3 px-5 py-3 hover:bg-accent/50 transition-colors duration-150 text-left"
                    @click="selectSource(sub.id)"
                  >
                    <Rss class="w-4 h-4 text-blue-500 flex-shrink-0" />
                    <div class="flex-1">
                      <p class="text-sm font-medium text-foreground">{{ sub.name }}</p>
                      <p class="text-xs text-muted-foreground">{{ sub.nodes.length }} nodes</p>
                    </div>
                    <CheckCircle v-if="appStore.selectedSubID === sub.id" class="w-4 h-4 text-green-500" />
                  </button>
                  <div class="border-t border-border ml-10" />
                </template>
              </template>
            </div>

            <div class="border-t border-border" />
            <div class="flex justify-end px-5 py-4">
              <button
                class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150"
                @click="showSourcePicker = false"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Export Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showExport && exportNode"
          class="fixed inset-0 z-50 flex items-center justify-center"
          @keydown.esc="showExport = false"
        >
          <div class="fixed inset-0 bg-black/50 backdrop-blur-sm" @click="showExport = false" />
          <div class="relative z-10 bg-card border border-border rounded-2xl shadow-lg flex flex-col overflow-hidden w-[520px] max-h-[85vh]">
            <div class="flex flex-col items-center pt-7 pb-5 px-6">
              <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3 bg-purple-500">
                <Share class="w-7 h-7 text-white" :stroke-width="1.5" />
              </div>
              <h2 class="text-lg font-bold text-card-foreground">Export Config</h2>
            </div>
            <div class="border-t border-border" />

            <!-- Tab Selector -->
            <div class="flex items-center gap-1 px-6 pt-4">
              <button
                v-for="tab in exportTabs"
                :key="tab.key"
                class="h-8 px-3 rounded-md text-xs font-medium transition-all duration-150"
                :class="exportTab === tab.key
                  ? 'bg-muted text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'"
                @click="exportTab = tab.key"
              >
                {{ tab.label }}
              </button>
            </div>

            <div class="px-6 py-4 space-y-4">
              <div class="space-y-1.5">
                <label class="text-xs font-medium text-muted-foreground">{{ exportTab === 'url' ? 'Config URL' : 'Config JSON' }}</label>
                <div class="p-3 rounded-lg bg-muted/50 text-xs font-mono text-muted-foreground break-all select-all max-h-32 overflow-y-auto whitespace-pre-wrap">
                  {{ exportTab === 'url' ? exportURL : exportJSON }}
                </div>
              </div>
              <div class="flex items-center gap-2 p-2.5 rounded-lg bg-blue-500/5 border border-blue-500/10">
                <Info class="w-4 h-4 text-blue-500 flex-shrink-0" />
                <span class="text-xs text-muted-foreground">
                  {{ exportTab === 'url' ? 'Copy this URL to import in other clients.' : 'Copy this JSON to import in other clients.' }}
                </span>
              </div>
            </div>
            <div class="border-t border-border" />
            <div class="flex items-center justify-between px-6 py-4">
              <button class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150" @click="showExport = false">
                Close
              </button>
              <button
                class="h-9 px-5 rounded-lg text-sm font-semibold text-white transition-all duration-150"
                :class="exportCopied ? 'bg-emerald-600' : 'bg-purple-600 hover:bg-purple-700'"
                @click="copyExport"
              >
                <span class="inline-flex items-center gap-2">
                  <Check v-if="exportCopied" class="w-4 h-4" />
                  <Copy v-else class="w-4 h-4" />
                  {{ exportCopied ? 'Copied!' : 'Copy' }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- IP Detail Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showIPDetail"
          class="fixed inset-0 z-50 flex items-center justify-center"
          @keydown.esc="showIPDetail = false"
        >
          <div class="fixed inset-0 bg-black/50 backdrop-blur-sm" @click="showIPDetail = false" />
          <div class="relative z-10 bg-card border border-border rounded-2xl shadow-lg flex flex-col overflow-hidden w-[480px] h-[500px]">
            <div class="flex flex-col items-center pt-7 pb-4 px-6">
              <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3 bg-indigo-500">
                <Globe class="w-7 h-7 text-white" :stroke-width="1.5" />
              </div>
              <h2 class="text-lg font-bold text-card-foreground">IP Information</h2>
            </div>
            <div class="border-t border-border" />

            <div class="flex-1 overflow-y-auto">
              <template v-if="ipInfoStore.loading && !ipInfoStore.info">
                <div class="flex flex-col items-center justify-center h-full gap-3">
                  <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
                  <p class="text-sm text-muted-foreground">Fetching IP info...</p>
                </div>
              </template>
              <template v-else-if="ipInfoStore.info">
                <div class="py-2">
                  <div
                    v-for="(item, i) in ipDetailItems"
                    :key="item[0]"
                  >
                    <div class="flex items-start px-6 py-2.5">
                      <span class="text-sm text-muted-foreground w-[120px] flex-shrink-0">{{ item[0] }}</span>
                      <span class="text-sm text-foreground break-all">{{ item[1] }}</span>
                    </div>
                    <div v-if="i < ipDetailItems.length - 1" class="border-t border-border ml-[148px]" />
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="flex items-center justify-center h-full">
                  <p class="text-sm text-muted-foreground">No data available</p>
                </div>
              </template>
            </div>

            <div class="border-t border-border" />
            <div class="flex items-center justify-between px-6 py-4">
              <span class="text-xs text-muted-foreground/50">Rate limit: 60 req/min</span>
              <div class="flex items-center gap-2">
                <button
                  class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150"
                  @click="showIPDetail = false"
                >
                  Close
                </button>
                <button
                  class="h-9 px-4 rounded-lg bg-blue-600 text-sm font-semibold text-white hover:bg-blue-700 transition-colors duration-150 inline-flex items-center gap-1.5"
                  :disabled="ipInfoStore.loading"
                  @click="ipInfoStore.fetch()"
                >
                  <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': ipInfoStore.loading }" />
                  Update
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useConnectionStore } from '@/stores/connection'
import { useIPInfoStore } from '@/stores/ipinfo'
import { api } from '@/api/client'
import { configURL, configJSON, protocolDisplayName, delayColorClass, delayString } from '@/lib/nodes'
import { countryFlag, type Node } from '@/types'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import UrlTestProgress from '@/components/ui/UrlTestProgress.vue'
import {
  Zap,
  Server,
  Globe,
  Share,
  Copy,
  Check,
  CheckCircle,
  Info,
  ChevronRight,
  AlertTriangle,
  AlertCircle,
  Settings,
  Loader2,
  RefreshCw,
  Radio,
  User,
  Inbox,
  Rss,
} from 'lucide-vue-next'

const appStore = useAppStore()
const connectionStore = useConnectionStore()
const ipInfoStore = useIPInfoStore()

// --- Resource Status ---
const resourceReady = ref(true)
onMounted(async () => {
  const { data } = await api.getResourceStatus()
  if (data) {
    resourceReady.value = data.xrayExists && data.geoIPExists && data.geoSiteExists
  }
})

// --- Unsaved Changes ---
const hasUnsavedChanges = ref(false)
const isReloading = ref(false)

// --- Selected Node ---
const selectedNode = computed(() => {
  if (!appStore.selectedNodeID) return null
  if (appStore.selectedSubID) {
    const sub = appStore.subscriptions.find(s => s.id === appStore.selectedSubID)
    return sub?.nodes.find(n => n.id === appStore.selectedNodeID) ?? null
  }
  return appStore.localNodes.find(n => n.id === appStore.selectedNodeID) ?? null
})

// --- Status ---
const statusLabel = computed(() => {
  switch (connectionStore.state) {
    case 'connected': return 'Connected'
    case 'connecting': return 'Connecting...'
    case 'disconnecting': return 'Disconnecting...'
    case 'error': return 'Error'
    default: return 'Disconnected'
  }
})

// --- Connect ---
function handleConnect() {
  if (connectionStore.isConnected) {
    connectionStore.disconnect()
  } else if (connectionStore.canConnect) {
    connectionStore.connect()
  }
}

function saveAndReload() {
  // Auto-save handles this via the debounced watcher
}

// --- System Proxy Toggle ---
async function toggleSystemProxy() {
  appStore.settings.enableSystemProxy = !appStore.settings.enableSystemProxy
  // Auto-save handles this via the debounced watcher
}

// --- URL Test ---
const urlTestRunning = ref(false)
const urlTestTested = ref(0)
const urlTestTotal = ref(0)

function handleURLTest() {
  if (urlTestRunning.value) return
  const nodes = sortedNodes.value
  if (nodes.length === 0) return
  urlTestTested.value = 0
  urlTestTotal.value = nodes.length
  urlTestRunning.value = true
  // Pass exact nodes to backend — no stale storage reads
  api.urlTestNodes(nodes as any)
}

function cancelURLTest() {
  urlTestRunning.value = false
}

onMounted(() => {
  EventsOn('urltest:result', (data: string) => {
    try {
      const result = JSON.parse(data) as { nodeId: string; delay: number | null }
      urlTestTested.value++
      const delay = result.delay ?? -1

      // Search all sources for the node by ID
      for (const sub of appStore.subscriptions) {
        const node = sub.nodes.find(n => n.id === result.nodeId)
        if (node) {
          node.delay = delay
          appStore.subscriptions = [...appStore.subscriptions]
          return
        }
      }
      const localNode = appStore.localNodes.find(n => n.id === result.nodeId)
      if (localNode) {
        localNode.delay = delay
        appStore.localNodes = [...appStore.localNodes]
      }
    } catch (e) {
      console.error('[URLTest] Failed to process result:', e)
    }
  })
  EventsOn('urltest:complete', () => {
    urlTestRunning.value = false
  })
})

onUnmounted(() => {
  EventsOff('urltest:result')
  EventsOff('urltest:complete')
})

// --- Update Sub ---
function handleUpdateSub() {
  if (appStore.selectedSubID) {
    // Verify the subscription still exists before refreshing
    const sub = appStore.subscriptions.find(s => s.id === appStore.selectedSubID)
    if (sub) {
      appStore.refreshSubscription(sub.id)
    } else {
      // selectedSubID is stale — try to re-match by finding any subscription
      if (appStore.subscriptions.length > 0) {
        appStore.refreshSubscription(appStore.subscriptions[0].id)
      }
    }
  }
}

// --- Node List ---
const maxVisibleNodes = 30
const showAllNodes = ref(false)

const sortedNodes = computed(() => {
  const nodes = appStore.selectedSubID
    ? appStore.subscriptions.find(s => s.id === appStore.selectedSubID)?.nodes || []
    : appStore.localNodes
  return [...nodes].sort((a, b) => {
    if (a.delay === null && b.delay === null) return 0
    if (a.delay === null) return 1
    if (b.delay === null) return -1
    if (a.delay === -1 && b.delay === -1) return 0
    if (a.delay === -1) return 1
    if (b.delay === -1) return -1
    return a.delay - b.delay
  })
})

const visibleNodes = computed(() =>
  showAllNodes.value ? sortedNodes.value : sortedNodes.value.slice(0, maxVisibleNodes)
)

// --- Source Picker ---
const showSourcePicker = ref(false)

const sourceLabel = computed(() => {
  if (!appStore.selectedSubID) return 'Local Nodes'
  const sub = appStore.subscriptions.find(s => s.id === appStore.selectedSubID)
  return sub?.name || 'Choose Source'
})

function selectSource(subID: string | null) {
  // Clear selection first via store method (persists to backend)
  appStore.setSelectedNodeID('')
  appStore.setSelectedSubID(subID || '')
  showSourcePicker.value = false
}

// --- Export ---
const showExport = ref(false)
const exportNode = ref<Node | null>(null)
const exportCopied = ref(false)
const exportTab = ref<'url' | 'json'>('url')
const exportTabs = [{ key: 'url' as const, label: 'URL' }, { key: 'json' as const, label: 'JSON' }]
let exportTimer: ReturnType<typeof setTimeout> | null = null

const exportURL = computed(() => exportNode.value ? configURL(exportNode.value) : '')
const exportJSON = computed(() => exportNode.value ? configJSON(exportNode.value) : '')

function openExport(node: Node) {
  exportNode.value = node
  exportCopied.value = false
  exportTab.value = 'url'
  showExport.value = true
}

async function copyExport() {
  try {
    const text = exportTab.value === 'url' ? exportURL.value : exportJSON.value
    await navigator.clipboard.writeText(text)
    exportCopied.value = true
    if (exportTimer) clearTimeout(exportTimer)
    exportTimer = setTimeout(() => { showExport.value = false }, 1000)
  } catch {}
}

// --- IP Detail ---
const showIPDetail = ref(false)

const ipDetailItems = computed(() => {
  const info = ipInfoStore.info
  if (!info) return []
  return [
    ['IP Address', info.ipAddress || '--'],
    ['Country', info.countryName || '--'],
    ['City', info.cityName || '--'],
    ['Region', info.regionName || '--'],
    ['Continent', info.continent || '--'],
    ['ASN', info.asn || '--'],
    ['Organization', info.asnOrganization || '--'],
    ['Timezone', info.timeZones?.join(', ') || '--'],
    ['Zip Code', info.zipCode || '--'],
    ['Latitude', String(info.latitude || '--')],
    ['Longitude', String(info.longitude || '--')],
    ['Proxy', info.isProxy ? 'Yes' : 'No'],
  ]
})
</script>
