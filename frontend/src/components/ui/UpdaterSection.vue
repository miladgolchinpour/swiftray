<template>
  <div>
    <!-- Downloading -->
    <template v-if="updaterStore.isDownloading">
      <!-- Progress Bar -->
      <div class="py-3 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium text-foreground">{{ updaterStore.status }}</span>
          <span class="text-xs text-muted-foreground tabular-nums">{{ Math.round(updaterStore.progress * 100) }}%</span>
        </div>
        <div class="h-2 rounded-full bg-muted overflow-hidden">
          <div class="h-full rounded-full bg-gradient-to-r from-orange-500 to-red-500 transition-all duration-300" :style="{ width: `${updaterStore.progress * 100}%` }" />
        </div>
        <div class="grid grid-cols-2 gap-x-4 gap-y-1 text-[11px]">
          <div v-if="updaterStore.url"><span class="text-muted-foreground">URL:</span> <span class="font-mono text-foreground truncate block">{{ updaterStore.url }}</span></div>
          <div v-if="updaterStore.destPath"><span class="text-muted-foreground">Dest:</span> <span class="font-mono text-foreground truncate block">{{ updaterStore.destPath }}</span></div>
          <div v-if="updaterStore.total > 0"><span class="text-muted-foreground">Downloaded:</span> <span class="font-mono text-foreground">{{ formatBytes(updaterStore.downloaded) }} / {{ formatBytes(updaterStore.total) }}</span></div>
          <div v-if="updaterStore.speed > 0"><span class="text-muted-foreground">Speed:</span> <span class="font-mono text-foreground">{{ formatSpeed(updaterStore.speed) }}</span></div>
          <div v-if="updaterStore.eta > 0"><span class="text-muted-foreground">ETA:</span> <span class="font-mono text-foreground">{{ formatETA(updaterStore.eta) }}</span></div>
          <div v-if="updaterStore.platform"><span class="text-muted-foreground">Platform:</span> <span class="font-mono text-foreground">{{ updaterStore.platform }}</span></div>
          <div v-if="updaterStore.currentVersion"><span class="text-muted-foreground">Version:</span> <span class="font-mono text-foreground">v{{ updaterStore.currentVersion }}</span></div>
          <div v-if="updaterStore.targetVersion"><span class="text-muted-foreground">Target:</span> <span class="font-mono text-blue-500">v{{ updaterStore.targetVersion }}</span></div>
        </div>
        <button class="w-full h-8 rounded-lg border border-border bg-card text-xs font-medium text-card-foreground hover:bg-accent transition-colors" @click="updaterStore.cancelDownload()">Cancel</button>
      </div>
      <!-- Log -->
      <div v-if="updaterStore.log.length > 0" class="mt-2 rounded-lg bg-muted/30 border border-border max-h-32 overflow-y-auto">
        <div class="p-2 space-y-0.5">
          <div v-for="(entry, i) in updaterStore.log" :key="i" class="flex gap-2 text-[10px] font-mono">
            <span class="text-muted-foreground/50 flex-shrink-0">{{ entry.time }}</span>
            <span :class="entry.type === 'error' ? 'text-red-500' : entry.type === 'success' ? 'text-green-500' : 'text-foreground/70'">{{ entry.message }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- Completed -->
    <template v-else-if="updaterStore.completed && updaterStore.completedMessage">
      <div class="py-3 space-y-2">
        <div class="p-3 rounded-lg bg-green-500/5 border border-green-500/10 space-y-2">
          <div class="flex items-center gap-2">
            <CheckCircle class="w-4 h-4 text-green-500" />
            <span class="text-xs font-medium text-green-600 dark:text-green-400">Runtime updated successfully.</span>
          </div>
          <div class="text-[11px] text-muted-foreground space-y-0.5">
            <p>Updated files:</p>
            <p>✓ xray</p>
            <p>✓ geoip.dat</p>
            <p>✓ geosite.dat</p>
          </div>
        </div>
        <p class="text-xs text-muted-foreground">A new runtime has been installed. Reload to apply?</p>
        <div class="flex gap-2">
          <button class="flex-1 h-9 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-colors" @click="updaterStore.resetState()">Later</button>
          <button class="flex-1 h-9 rounded-lg bg-blue-600 text-sm font-semibold text-white hover:bg-blue-700 transition-colors inline-flex items-center justify-center gap-2" @click="reloadRuntime">
            <RefreshCw class="w-4 h-4" />Reload Now
          </button>
        </div>
      </div>
    </template>

    <!-- Error -->
    <template v-else-if="updaterStore.hasError">
      <div class="py-2 flex items-center gap-2">
        <XCircle class="w-4 h-4 text-red-500" />
        <span class="text-xs text-red-500">{{ updaterStore.lastError }}</span>
      </div>
      <button class="w-full h-8 rounded-lg border border-border bg-card text-xs font-medium text-card-foreground hover:bg-accent transition-colors mt-2" @click="updaterStore.resetState()">Dismiss</button>
    </template>

    <!-- Check for Updates (default) -->
    <template v-else>
      <button :disabled="checking" class="w-full h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-all duration-150 disabled:opacity-50 inline-flex items-center justify-center gap-2 mt-1" @click="checkUpdate">
        <Loader2 v-if="checking" class="w-4 h-4 animate-spin" />
        <RefreshCw v-else class="w-4 h-4" />
        {{ checking ? 'Checking...' : 'Check for Updates' }}
      </button>
      <template v-if="updateStatus?.hasUpdate">
        <div class="mt-2 p-3 rounded-lg bg-blue-500/5 border border-blue-500/10 space-y-2">
          <div class="flex items-center gap-2">
            <ArrowDownCircle class="w-4 h-4 text-blue-500" />
            <span class="text-xs font-medium text-foreground">Update available</span>
          </div>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div><span class="text-muted-foreground">Installed:</span> <span class="font-mono">{{ updaterStore.currentVersion || 'unknown' }}</span></div>
            <div><span class="text-muted-foreground">Available:</span> <span class="font-mono text-blue-500">v{{ updateStatus.latestVersion }}</span></div>
            <div><span class="text-muted-foreground">Release:</span> <span class="font-mono">{{ formatDate(updateStatus.releaseDate) }}</span></div>
            <div><span class="text-muted-foreground">Platform:</span> <span class="font-mono">{{ platformDisplay }}</span></div>
            <div class="col-span-2"><span class="text-muted-foreground">Files:</span> <span class="font-mono">xray, geoip.dat, geosite.dat</span></div>
          </div>
        </div>
        <button class="w-full h-10 rounded-lg mt-2 bg-orange-600 text-sm font-semibold text-white hover:bg-orange-700 transition-colors inline-flex items-center justify-center gap-2" @click="startDownload">
          <Download class="w-4 h-4" />Download & Install
        </button>
      </template>
      <p v-else-if="updateStatus && !updateStatus.hasUpdate && !updateStatus.error" class="text-xs text-green-500 flex items-center gap-1.5 mt-1.5">
        <CheckCircle class="w-3.5 h-3.5" />You're up to date.
      </p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useUpdaterStore } from '@/stores/updater'
import { useConnectionStore } from '@/stores/connection'
import { useNotifications } from '@/composables/useNotifications'
import { api } from '@/api/client'
import { CheckCircle, XCircle, Loader2, RefreshCw, Download, ArrowDownCircle } from 'lucide-vue-next'

const updaterStore = useUpdaterStore()
const connectionStore = useConnectionStore()
const notify = useNotifications()

const checking = ref(false)
const updateStatus = ref<any>(null)

const platformDisplay = computed(() => { const ua = navigator.userAgent; if (ua.includes('Mac')) return 'macOS'; if (ua.includes('Win')) return 'Windows'; if (ua.includes('Linux')) return 'Linux'; return 'Unknown' })

async function checkUpdate() {
  checking.value = true
  updateStatus.value = null
  const { data, error } = await updaterStore.checkUpdate()
  updateStatus.value = error ? { error } : data
  checking.value = false
}

async function startDownload() {
  await updaterStore.download()
}

async function reloadRuntime() {
  notify.info('Reloading runtime...')
  const wasConnected = connectionStore.isConnected
  if (wasConnected) {
    const { error } = await api.saveAndReload((await api.getSettings()).data as any)
    if (error) notify.error('Reload failed', error)
    else notify.success('Runtime reloaded')
  }
  updaterStore.resetState()
}

function formatBytes(b: number): string {
  if (b < 1024) return `${b} B`
  if (b < 1048576) return `${(b / 1024).toFixed(1)} KB`
  return `${(b / 1048576).toFixed(1)} MB`
}

function formatSpeed(bps: number): string {
  if (bps < 1048576) return `${(bps / 1024).toFixed(1)} KB/s`
  return `${(bps / 1048576).toFixed(1)} MB/s`
}

function formatETA(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

function formatDate(d: string): string {
  if (!d) return ''
  return new Date(d).toLocaleDateString()
}
</script>
