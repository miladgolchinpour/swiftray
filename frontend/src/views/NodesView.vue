<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <!-- Header -->
        <div class="space-y-1">
          <h1 class="text-[28px] font-bold tracking-tight text-foreground">Local Nodes</h1>
          <p class="text-sm text-muted-foreground">Manually added proxy nodes</p>
        </div>

        <!-- Action Buttons -->
        <div class="grid grid-cols-2 gap-3">
          <button
            :disabled="displayedNodes.length === 0 || urlTestRunning"
            class="h-10 rounded-lg bg-blue-600 text-sm font-semibold text-white hover:bg-blue-700 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2"
            @click="startURLTest"
          >
            <Radio class="w-4 h-4" :class="{ 'animate-pulse': urlTestRunning }" />
            {{ urlTestRunning ? 'Testing...' : 'URL Test' }}
          </button>
          <button
            :disabled="displayedNodes.length === 0"
            class="h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2"
            @click="sortByPing = !sortByPing"
          >
            <ArrowUpDown v-if="sortByPing" class="w-4 h-4" />
            <ArrowDown v-else class="w-4 h-4" />
            {{ sortByPing ? 'Sort by Name' : 'Sort by Ping' }}
          </button>
        </div>

        <!-- URL Test Progress -->
        <UrlTestProgress
          v-if="urlTestRunning"
          :tested="urlTestTested"
          :total="urlTestTotal"
          @cancel="cancelURLTest"
        />

        <!-- Local Nodes Section -->
        <SectionCard
          title="Local Nodes"
          :icon="Server"
          icon-bg-class="bg-purple-500"
          :count="store.localNodes.length"
          :collapsed="collapsed"
          @update:collapsed="collapsed = $event"
        >
          <template #actions>
            <button
              v-if="store.localNodes.length > 0"
              class="h-7 px-3 rounded-md border border-border bg-card text-xs font-semibold text-red-500 hover:bg-red-50 hover:text-red-600 hover:border-red-200 dark:hover:bg-red-950 dark:hover:text-red-400 dark:hover:border-red-800 transition-all duration-150 inline-flex items-center gap-1.5"
              @click.stop="showDeleteAll = true"
            >
              <Trash2 class="w-3 h-3" />
              Delete All
            </button>
            <button
              class="h-7 px-3 rounded-md bg-purple-600 text-xs font-semibold text-white hover:bg-purple-700 transition-all duration-150 inline-flex items-center gap-1.5"
              @click.stop="openAddSheet"
            >
              <Plus class="w-3 h-3" />
              Add
            </button>
          </template>

          <!-- Empty State -->
          <div v-if="store.localNodes.length === 0" class="flex flex-col items-center py-10 gap-3">
            <Server class="w-9 h-9 text-muted-foreground/30" />
            <p class="text-sm font-semibold text-muted-foreground">No Local Nodes</p>
            <p class="text-xs text-muted-foreground/70">Add nodes manually to use them as proxy</p>
          </div>

          <!-- Node List -->
          <div v-else>
            <div
              v-for="node in displayedNodes"
              :key="node.id"
              class="flex items-center gap-2.5 px-[18px] py-3 cursor-pointer select-none transition-colors duration-150 hover:bg-accent/50"
              :class="{ 'bg-accent/30': store.selectedNodeID === node.id }"
              @click="selectNode(node.id)"
            >
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-card-foreground truncate">{{ node.name }}</p>
                <div class="flex items-center gap-1.5 mt-0.5">
                  <span class="inline-flex items-center h-5 px-2 rounded-full bg-muted text-[10px] font-mono text-muted-foreground">
                    {{ protocolDisplayName(node.protocolType) }}
                  </span>
                  <span class="text-xs text-muted-foreground">{{ node.address }}:{{ node.port }}</span>
                  <span class="text-xs font-mono" :class="delayColorClass(node.delay)">
                    {{ delayString(node.delay) }}
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-1.5 flex-shrink-0">
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-accent hover:text-accent-foreground"
                  title="Export config"
                  @click.stop="openExport(node)"
                >
                  <Share class="w-3.5 h-3.5" />
                </button>
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-accent hover:text-accent-foreground"
                  title="Edit node"
                  @click.stop="openEdit(node)"
                >
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-red-50 hover:text-red-600 hover:border-red-200 dark:hover:bg-red-950 dark:hover:text-red-400 dark:hover:border-red-800"
                  title="Delete node"
                  @click.stop="confirmDelete(node)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </SectionCard>
      </div>
    </div>

    <!-- ==================== ADD MODAL ==================== -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showAdd"
          class="fixed inset-0 z-50 flex items-center justify-center"
          @keydown.esc="closeAddSheet"
        >
          <div class="fixed inset-0 bg-black/50 backdrop-blur-sm" @click="closeAddSheet" />
          <div class="relative z-10 bg-card border border-border rounded-2xl shadow-lg flex flex-col overflow-hidden w-[520px] max-h-[85vh]">
            <!-- Header -->
            <div class="flex flex-col items-center pt-7 pb-5 px-6">
              <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3 bg-purple-500">
                <PlusCircle class="w-7 h-7 text-white" :stroke-width="1.5" />
              </div>
              <h2 class="text-lg font-bold text-card-foreground">
                {{ addStep === 0 ? 'Add Local Node' : 'Verify Config' }}
              </h2>
            </div>
            <div class="border-t border-border" />

            <!-- Step 0: Choose Method -->
            <template v-if="addStep === 0">
              <div class="flex-1 overflow-y-auto px-6 py-5 space-y-3">
                <button
                  v-for="method in addMethods"
                  :key="method.title"
                  class="w-full flex items-center gap-3.5 p-3.5 rounded-xl bg-muted/50 hover:bg-muted transition-colors duration-150 text-left"
                  @click="method.action"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0" :class="method.color">
                    <component :is="method.icon" class="w-5 h-5 text-white" :stroke-width="2" />
                  </div>
                  <span class="text-sm font-medium text-card-foreground">{{ method.title }}</span>
                  <ChevronRight class="w-4 h-4 text-muted-foreground ml-auto" />
                </button>
                <p v-if="addError" class="text-xs text-red-500 flex items-center gap-1.5 mt-2">
                  <AlertTriangle class="w-3.5 h-3.5 flex-shrink-0" />
                  {{ addError }}
                </p>
              </div>
              <div class="border-t border-border" />
              <div class="flex items-center justify-between px-6 py-4">
                <button class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150" @click="closeAddSheet">
                  Cancel
                </button>
              </div>
            </template>

            <!-- Step 1: Verify Imported Config -->
            <template v-if="addStep === 1">
              <div class="flex-1 overflow-y-auto px-6 py-5 space-y-4">
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-muted-foreground">Imported Config</label>
                  <div class="p-2.5 rounded-lg bg-muted/50 text-xs font-mono text-muted-foreground max-h-20 overflow-y-auto break-all select-all">
                    {{ addRawText }}
                  </div>
                </div>
                <div class="border-t border-border" />
                <div class="grid grid-cols-[1fr_auto] gap-2 items-end">
                  <FormField label="Name">
                    <input v-model="addDraft.name" type="text" placeholder="Node Name" class="field-input" />
                  </FormField>
                  <span class="inline-flex items-center h-9 px-3 rounded-full bg-muted text-xs font-mono text-muted-foreground">
                    {{ protocolDisplayName(addDraft.protocolType) }}
                  </span>
                </div>
                <div class="grid grid-cols-[1fr_100px] gap-2">
                  <FormField label="Address">
                    <input v-model="addDraft.address" type="text" placeholder="example.com" class="field-input" />
                  </FormField>
                  <FormField label="Port">
                    <input v-model.number="addDraft.port" type="number" placeholder="443" class="field-input" />
                  </FormField>
                </div>
                <div class="flex items-center gap-2">
                  <span class="inline-flex items-center h-6 px-2.5 rounded-full bg-muted text-[10px] font-mono text-muted-foreground">
                    {{ addDraft.transport }}
                  </span>
                  <span
                    class="inline-flex items-center h-6 px-2.5 rounded-full text-[10px] font-medium"
                    :class="addDraft.tls ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400' : 'bg-red-500/15 text-red-600 dark:text-red-400'"
                  >
                    {{ addDraft.tls ? 'TLS' : 'No TLS' }}
                  </span>
                </div>
                <FormField label="Fingerprint">
                  <select v-model="addDraft.fingerprint" class="field-input">
                    <option v-for="f in fingerprints" :key="f" :value="f">{{ f }}</option>
                  </select>
                </FormField>
              </div>
              <div class="border-t border-border" />
              <div class="flex items-center justify-between px-6 py-4">
                <button class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150" @click="addStep = 0">
                  Back
                </button>
                <button
                  class="h-9 px-5 rounded-lg text-sm font-semibold text-white bg-purple-600 hover:bg-purple-700 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
                  :disabled="!addDraft.name || !addDraft.address"
                  @click="submitAddImported"
                >
                  Add Node
                </button>
              </div>
            </template>

            <!-- Step 2: Manual Form -->
            <template v-if="addStep === 2">
              <div class="flex-1 overflow-y-auto px-6 py-5">
                <NodeForm :model="addDraft" @update:model="addDraft = $event" />
              </div>
              <div class="border-t border-border" />
              <div class="flex items-center justify-between px-6 py-4">
                <button class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent transition-colors duration-150" @click="addStep = 0">
                  Back
                </button>
                <button
                  class="h-9 px-5 rounded-lg text-sm font-semibold text-white bg-purple-600 hover:bg-purple-700 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
                  :disabled="!addDraft.name || !addDraft.address"
                  @click="submitAddManual"
                >
                  Add Node
                </button>
              </div>
            </template>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ==================== EDIT MODAL ==================== -->
    <ModalSheet
      v-model="showEdit"
      title="Edit Node"
      :icon="Pencil"
      icon-bg-class="bg-purple-500"
      action-label="Save"
      action-btn-class="bg-purple-600 hover:bg-purple-700"
      :disabled="!editDraft.name || !editDraft.address"
      :loading="editLoading"
      :width="520"
      @action="submitEdit"
    >
      <NodeForm :model="editDraft" @update:model="editDraft = $event" />
    </ModalSheet>

    <!-- ==================== EXPORT MODAL ==================== -->
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

    <!-- ==================== DELETE CONFIRM ==================== -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      title="Delete Node?"
      :message="deleteTarget ? `This will remove '${deleteTarget.name}'. This cannot be undone.` : ''"
      confirm-label="Delete"
      :loading="deleteLoading"
      @confirm="handleDelete"
    />

    <!-- ==================== DELETE ALL CONFIRM ==================== -->
    <ConfirmDialog
      v-model="showDeleteAll"
      title="Delete All Local Nodes?"
      :message="`This will remove all ${store.localNodes.length} local nodes. This cannot be undone.`"
      confirm-label="Delete All"
      :loading="deleteLoading"
      @confirm="handleDeleteAll"
    />

    <!-- ==================== BULK IMPORT CONFIRM ==================== -->
    <ConfirmDialog
      v-model="showBulkImport"
      :title="`${bulkImportNodes.length} Nodes Found`"
      :message="`Import ${bulkImportNodes.length} valid node${bulkImportNodes.length === 1 ? '' : 's'}? Invalid entries were skipped.`"
      confirm-label="Import All"
      @confirm="handleBulkImport"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useNotifications } from '@/composables/useNotifications'
import { configURL, configJSON, protocolDisplayName, delayColorClass, delayString } from '@/lib/nodes'
import type { Node, NodeProtocol } from '@/types'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import SectionCard from '@/components/ui/SectionCard.vue'
import ModalSheet from '@/components/ui/ModalSheet.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import UrlTestProgress from '@/components/ui/UrlTestProgress.vue'
import NodeForm from '@/components/ui/NodeForm.vue'
import FormField from '@/components/ui/FormField.vue'
import {
  Server,
  Plus,
  PlusCircle,
  Pencil,
  Trash2,
  Share,
  Copy,
  Check,
  Info,
  ChevronRight,
  AlertTriangle,
  Radio,
  ArrowUpDown,
  ArrowDown,
  Keyboard,
  ClipboardPaste,
  FileJson,
} from 'lucide-vue-next'

const store = useAppStore()
const notify = useNotifications()

// --- Sorting ---
const sortByPing = ref(false)
const collapsed = ref(false)

const displayedNodes = computed(() => {
  const nodes = store.localNodes
  if (!sortByPing.value) return nodes
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

// --- Node Selection ---
function selectNode(id: string) {
  store.setSelectedNodeID(id)
}

// --- URL Test ---
const urlTestRunning = ref(false)
const urlTestTested = ref(0)
const urlTestTotal = ref(0)
const urlTestResults = ref<Map<string, number | null>>(new Map())

function startURLTest() {
  if (urlTestRunning.value) return
  urlTestResults.value.clear()
  urlTestTested.value = 0
  urlTestTotal.value = store.localNodes.length
  urlTestRunning.value = true
  api.urlTestLocal()
}

function cancelURLTest() {
  urlTestRunning.value = false
  urlTestTested.value = 0
}

import { api } from '@/api/client'

onMounted(() => {
  EventsOn('urltest:result', (data: string) => {
    try {
      const result = JSON.parse(data) as { nodeId: string; delay: number | null }
      urlTestResults.value.set(result.nodeId, result.delay)
      urlTestTested.value = urlTestResults.value.size

      // Update node delay in store
      const node = store.localNodes.find(n => n.id === result.nodeId)
      if (node) {
        node.delay = result.delay ?? -1
      }
    } catch {}
  })

  EventsOn('urltest:complete', () => {
    urlTestRunning.value = false
  })
})

onUnmounted(() => {
  EventsOff('urltest:result')
  EventsOff('urltest:complete')
})

// --- Add Node ---
const showAdd = ref(false)
const addStep = ref(0)
const addError = ref('')
const addRawText = ref('')
const addDraft = ref<Node>(emptyNode())
const bulkImportNodes = ref<Node[]>([])
const showBulkImport = ref(false)

const fingerprints = ['chrome', 'firefox', 'safari', 'edge', 'ios', 'android', 'random', 'randomized', 'none']

function emptyNode(): Node {
  return {
    id: '',
    name: '',
    address: '',
    port: 443,
    protocolType: 'vless',
    transport: 'tcp',
    tls: true,
    delay: null,
    uuid: '',
    password: '',
    alterId: 0,
    security: 'auto',
    cipher: 'aes-256-gcm',
    flow: '',
    encryption: 'none',
    sni: '',
    fingerprint: 'chrome',
    alpn: '',
    realityPublicKey: '',
    realityShortId: '',
    realitySpiderX: '',
    host: '',
    path: '',
    serviceName: '',
    serviceMode: '',
    rawLink: '',
  }
}

function openAddSheet() {
  addStep.value = 0
  addError.value = ''
  addRawText.value = ''
  addDraft.value = emptyNode()
  bulkImportNodes.value = []
  showAdd.value = true
}

function closeAddSheet() {
  showAdd.value = false
  addStep.value = 0
}

const addMethods = [
  { title: 'Enter Manually', icon: Keyboard, color: 'bg-blue-500', action: () => { addStep.value = 2 } },
  { title: 'Paste Config URL', icon: ClipboardPaste, color: 'bg-emerald-500', action: () => pasteURL() },
  { title: 'Paste Config JSON', icon: FileJson, color: 'bg-orange-500', action: () => pasteJSON() },
]

// --- Paste URL ---
async function pasteURL() {
  try {
    const text = await navigator.clipboard.readText()
    const clip = text.trim()
    if (!clip) {
      addError.value = 'Clipboard is empty'
      return
    }
    addRawText.value = clip

    const lines = clip.split('\n').map(l => l.trim()).filter(Boolean)
    const allNodes: Node[] = []

    for (const line of lines) {
      const node = parseConfigURL(line)
      if (node) allNodes.push(node)
    }

    if (allNodes.length === 0) {
      addError.value = 'No valid config URLs found in clipboard'
      addRawText.value = ''
      return
    }

    addError.value = ''
    if (allNodes.length === 1) {
      addDraft.value = { ...allNodes[0], id: '' }
      addStep.value = 1
    } else {
      bulkImportNodes.value = allNodes
      showBulkImport.value = true
    }
  } catch {
    addError.value = 'Could not read clipboard'
  }
}

// --- Paste JSON ---
async function pasteJSON() {
  try {
    const text = await navigator.clipboard.readText()
    const clip = text.trim()
    if (!clip) {
      addError.value = 'Clipboard is empty'
      return
    }

    const nodes = parseJSONNodes(clip)
    if (nodes.length === 0) {
      addError.value = 'No valid nodes found in JSON'
      return
    }

    addRawText.value = clip
    addError.value = ''

    if (nodes.length === 1) {
      addDraft.value = { ...nodes[0], id: '' }
      addStep.value = 1
    } else {
      bulkImportNodes.value = nodes
      showBulkImport.value = true
    }
  } catch {
    addError.value = 'Could not read clipboard'
  }
}

// --- Parse Config URL ---
function parseConfigURL(url: string): Node | null {
  try {
    if (url.startsWith('vmess://')) {
      return parseVMessURL(url)
    } else if (url.startsWith('vless://')) {
      return parseVLessURL(url)
    } else if (url.startsWith('trojan://')) {
      return parseTrojanURL(url)
    } else if (url.startsWith('ss://')) {
      return parseShadowsocksURL(url)
    }
  } catch {}
  return null
}

function parseVMessURL(url: string): Node | null {
  const base64 = url.slice(8)
  const json = JSON.parse(atob(base64))
  return {
    ...emptyNode(),
    name: json.ps || '',
    protocolType: 'vmess',
    address: json.add || '',
    port: parseInt(json.port) || 443,
    uuid: json.id || '',
    alterId: parseInt(json.aid) || 0,
    security: json.scy || 'auto',
    transport: (json.net || 'tcp') as any,
    tls: json.tls === 'tls',
    host: json.host || '',
    path: json.path || '',
    sni: json.sni || '',
    fingerprint: json.fp || 'chrome',
  }
}

function parseVLessURL(url: string): Node {
  const hashIdx = url.indexOf('#')
  const name = hashIdx > 0 ? decodeURIComponent(url.slice(hashIdx + 1)) : ''
  const withoutHash = hashIdx > 0 ? url.slice(0, hashIdx) : url
  const questionIdx = withoutHash.indexOf('?')
  const query = questionIdx > 0 ? withoutHash.slice(questionIdx + 1) : ''
  const authority = withoutHash.slice(8, questionIdx > 0 ? questionIdx : undefined)

  const atIndex = authority.lastIndexOf('@')
  const uuid = atIndex > 0 ? authority.slice(0, atIndex) : ''
  const hostPort = atIndex > 0 ? authority.slice(atIndex + 1) : authority
  const colonIdx = hostPort.lastIndexOf(':')
  const address = colonIdx > 0 ? hostPort.slice(0, colonIdx) : hostPort
  const port = parseInt(hostPort.slice(colonIdx + 1)) || 443

  const params = new URLSearchParams(query)
  return {
    ...emptyNode(),
    name,
    protocolType: 'vless',
    address,
    port,
    uuid,
    transport: (params.get('type') || 'tcp') as any,
    tls: params.get('security') === 'tls',
    sni: params.get('sni') || '',
    fingerprint: params.get('fp') || 'chrome',
    host: params.get('host') || '',
    path: params.get('path') || '',
    flow: params.get('flow') || '',
  }
}

function parseTrojanURL(url: string): Node {
  const hashIdx = url.indexOf('#')
  const name = hashIdx > 0 ? decodeURIComponent(url.slice(hashIdx + 1)) : ''
  const withoutHash = hashIdx > 0 ? url.slice(0, hashIdx) : url
  const questionIdx = withoutHash.indexOf('?')
  const query = questionIdx > 0 ? withoutHash.slice(questionIdx + 1) : ''
  const authority = withoutHash.slice(9, questionIdx > 0 ? questionIdx : undefined)

  const atIndex = authority.lastIndexOf('@')
  const password = atIndex > 0 ? authority.slice(0, atIndex) : ''
  const hostPort = atIndex > 0 ? authority.slice(atIndex + 1) : authority
  const colonIdx = hostPort.lastIndexOf(':')
  const address = colonIdx > 0 ? hostPort.slice(0, colonIdx) : hostPort
  const port = parseInt(hostPort.slice(colonIdx + 1)) || 443

  const params = new URLSearchParams(query)
  return {
    ...emptyNode(),
    name,
    protocolType: 'trojan',
    address,
    port,
    password,
    transport: (params.get('type') || 'tcp') as any,
    tls: params.get('security') === 'tls',
    sni: params.get('sni') || '',
    fingerprint: params.get('fp') || 'chrome',
    host: params.get('host') || '',
    path: params.get('path') || '',
  }
}

function parseShadowsocksURL(url: string): Node {
  const hashIdx = url.indexOf('#')
  const name = hashIdx > 0 ? decodeURIComponent(url.slice(hashIdx + 1)) : ''
  const withoutHash = hashIdx > 0 ? url.slice(0, hashIdx) : url

  let b64 = withoutHash.slice(5)
  // Fix base64 padding
  b64 = b64.replace(/-/g, '+').replace(/_/g, '/')
  while (b64.length % 4) b64 += '='

  const decoded = atob(b64)
  const atIndex = decoded.lastIndexOf('@')
  const cipher = atIndex > 0 ? decoded.slice(0, atIndex) : 'aes-256-gcm'
  const password = atIndex > 0 ? decoded.slice(atIndex + 1) : ''

  // Parse host:port from remaining URL
  const rest = withoutHash.slice(5 + decoded.length + (atIndex > 0 ? 1 : 0))
  // Actually, SS URL format is ss://base64@host:port#name
  // The base64 part is cipher:password, and host:port comes after @
  // But our parsing above already handled it. Let me re-parse.
  const afterSS = withoutHash.slice(5)
  const parts = afterSS.split('@')
  if (parts.length === 2) {
    const hostPort = parts[1]
    const colonIdx = hostPort.lastIndexOf(':')
    const address = colonIdx > 0 ? hostPort.slice(0, colonIdx) : hostPort
    const port = parseInt(hostPort.slice(colonIdx + 1)) || 443

    return {
      ...emptyNode(),
      name,
      protocolType: 'ss',
      address,
      port,
      cipher,
      password,
    }
  }

  return { ...emptyNode(), name, protocolType: 'ss', address: '', port: 443, cipher, password }
}

// --- Parse JSON ---
function parseJSONNodes(jsonStr: string): Node[] {
  try {
    const json = JSON.parse(jsonStr)
    let outbounds: any[] = []

    if (Array.isArray(json)) {
      outbounds = json
    } else if (json.outbounds && Array.isArray(json.outbounds)) {
      outbounds = json.outbounds
    }

    return outbounds
      .map(ob => parseJSONOutbound(ob))
      .filter((n): n is Node => n !== null && n.address !== '')
  } catch {
    return []
  }
}

function parseJSONOutbound(ob: any): Node | null {
  let proto: string = String(ob.protocol || 'vless').toLowerCase()
  if (proto === 'shadowsocks') proto = 'ss'
  let addr = '', prt = 443, uuidVal = '', pw = '', aid = 0
  let fl = '', cr = 'aes-256-gcm', tr = 'tcp', tl = true
  let sn = '', ho = '', pa = '', fp = 'chrome', sn2 = ''

  const settings = ob.settings || {}
  const vnext = settings.vnext?.[0]
  const servers = settings.servers?.[0]

  if (vnext) {
    addr = vnext.address || ''
    prt = vnext.port || 443
    const user = vnext.users?.[0]
    if (user) {
      uuidVal = user.id || ''
      fl = user.flow || ''
      aid = user.alterId || 0
    }
  }
  if (servers) {
    addr = addr || servers.address || ''
    prt = servers.port || prt
    pw = servers.password || ''
    cr = servers.method || 'aes-256-gcm'
  }

  if (!addr) return null

  const stream = ob.streamSettings || {}
  tr = String(stream.network || 'tcp').toLowerCase()
  const sec = String(stream.security || 'none').toLowerCase()
  tl = sec === 'tls' || sec === 'reality'
  if (stream.tlsSettings) {
    sn = stream.tlsSettings.serverName || ''
    if (stream.tlsSettings.fingerprint) fp = stream.tlsSettings.fingerprint
  }
  if (stream.wsSettings) {
    pa = stream.wsSettings.path || ''
    ho = stream.wsSettings.headers?.Host || ''
  }
  if (stream.grpcSettings) {
    sn2 = stream.grpcSettings.serviceName || ''
  }

  return {
    ...emptyNode(),
    name: `${protocolDisplayName(proto as NodeProtocol)} ${addr}`,
    protocolType: proto as NodeProtocol,
    address: addr,
    port: prt,
    transport: tr as any,
    tls: tl,
    uuid: uuidVal,
    password: pw,
    alterId: aid,
    cipher: proto === 'ss' ? cr : undefined as any,
    flow: fl,
    sni: sn,
    fingerprint: fp,
    host: ho,
    path: pa,
    serviceName: sn2,
  }
}

// --- Submit Add ---
function submitAddImported() {
  const node: Node = { ...addDraft.value, id: crypto.randomUUID() }
  store.addLocalNode(node)
  showAdd.value = false
}

function submitAddManual() {
  if (!addDraft.value.name || !addDraft.value.address) return
  const node: Node = { ...addDraft.value, id: crypto.randomUUID() }
  store.addLocalNode(node)
  showAdd.value = false
}

function handleBulkImport() {
  for (const n of bulkImportNodes.value) {
    store.addLocalNode({ ...n, id: crypto.randomUUID() })
  }
  showBulkImport.value = false
  showAdd.value = false
  notify.success(`${bulkImportNodes.value.length} nodes imported`)
}

// --- Edit Node ---
const showEdit = ref(false)
const editDraft = ref<Node>(emptyNode())
const editLoading = ref(false)

function openEdit(node: Node) {
  editDraft.value = { ...node }
  showEdit.value = true
}

async function submitEdit() {
  editLoading.value = true
  await store.updateLocalNode(editDraft.value)
  editLoading.value = false
  showEdit.value = false
}

// --- Export Node ---
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
  } catch {
    notify.error('Failed to copy', 'Could not access clipboard')
  }
}

// --- Delete ---
const showDeleteConfirm = ref(false)
const showDeleteAll = ref(false)
const deleteTarget = ref<Node | null>(null)
const deleteLoading = ref(false)

function confirmDelete(node: Node) {
  deleteTarget.value = node
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  await store.deleteLocalNode(deleteTarget.value.id)
  deleteLoading.value = false
  showDeleteConfirm.value = false
  deleteTarget.value = null
}

async function handleDeleteAll() {
  deleteLoading.value = true
  const ids = store.localNodes.map(n => n.id)
  for (const id of ids) {
    await store.deleteLocalNode(id)
  }
  deleteLoading.value = false
  showDeleteAll.value = false
}
</script>

<style scoped>
.field-input {
  @apply h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150;
}
</style>
