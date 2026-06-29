import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

export interface UpdaterLogEntry {
  time: string
  message: string
  type: 'info' | 'success' | 'error'
}

export const useUpdaterStore = defineStore('updater', () => {
  // --- State ---
  const xrayVersion = ref('')
  const platform = ref('')
  const arch = ref('')

  // Download state
  const downloading = ref(false)
  const stage = ref('')
  const status = ref('')
  const progress = ref(0)
  const url = ref('')
  const destPath = ref('')
  const downloaded = ref(0)
  const total = ref(0)
  const speed = ref(0)
  const eta = ref(0)
  const currentVersion = ref('')
  const targetVersion = ref('')
  const lastError = ref('')

  // Completion state (visible briefly)
  const completed = ref(false)
  const completedMessage = ref('')
  let completeTimer: ReturnType<typeof setTimeout> | null = null

  // Log
  const log = ref<UpdaterLogEntry[]>([])

  // --- Computed ---
  const isDownloading = computed(() => downloading.value)
  const hasError = computed(() => lastError.value !== '')

  // --- Event listener ---
  let listening = false

  function startListening() {
    if (listening) return
    listening = true

    EventsOn('updater:progress', (data: string) => {
      try {
        const p = JSON.parse(data)

        if (p.error) {
          downloading.value = false
          lastError.value = p.error
          addLog(p.error, 'error')
          return
        }

        if (p.stage === 'completed') {
          downloading.value = false
          progress.value = 1
          status.value = p.status || 'Completed'
          addLog(p.status || 'Completed', 'success')

          completed.value = true
          completedMessage.value = p.status || 'Completed'
          if (completeTimer) clearTimeout(completeTimer)
          completeTimer = setTimeout(() => {
            completed.value = false
            completedMessage.value = ''
          }, 5000)
          return
        }

        // Progress update
        downloading.value = true
        stage.value = p.stage || ''
        status.value = p.status || ''
        progress.value = p.progress || 0
        url.value = p.url || ''
        destPath.value = p.destPath || ''
        downloaded.value = p.downloaded || 0
        total.value = p.total || 0
        speed.value = p.speed || 0
        eta.value = p.eta || 0
        currentVersion.value = p.version || currentVersion.value
        targetVersion.value = p.targetVersion || targetVersion.value
        platform.value = p.platform || platform.value
        arch.value = p.arch || arch.value

        if (p.status && !log.value.some(e => e.message === p.status)) {
          addLog(p.status, 'info')
        }
      } catch {}
    })
  }

  function stopListening() {
    EventsOff('updater:progress')
    listening = false
  }

  function addLog(message: string, type: 'info' | 'success' | 'error' = 'info') {
    const now = new Date()
    const time = now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
    log.value.push({ time, message, type })
    if (log.value.length > 100) {
      log.value = log.value.slice(-100)
    }
  }

  function clearLog() {
    log.value = []
  }

  // --- Actions ---
  async function checkUpdate() {
    const { data, error } = await api.checkXrayUpdateStatus()
    if (error) return { error }
    return { data }
  }

  async function download() {
    if (downloading.value) return { error: 'Download already in progress' }
    clearLog()
    lastError.value = ''
    completed.value = false
    addLog('Starting runtime update...', 'info')
    const { error } = await api.downloadXrayUpdate()
    if (error) {
      lastError.value = error
      addLog(error, 'error')
      return { error }
    }
    return { ok: true }
  }

  async function cancelDownload() {
    await api.cancelDownload()
    downloading.value = false
    addLog('Download cancelled', 'info')
  }

  function resetState() {
    downloading.value = false
    stage.value = ''
    status.value = ''
    progress.value = 0
    url.value = ''
    destPath.value = ''
    downloaded.value = 0
    total.value = 0
    speed.value = 0
    eta.value = 0
    lastError.value = ''
    completed.value = false
    completedMessage.value = ''
  }

  return {
    xrayVersion, platform, arch,
    downloading, stage, status, progress,
    url, destPath, downloaded, total, speed, eta,
    currentVersion, targetVersion, lastError,
    completed, completedMessage,
    log,
    isDownloading, hasError,
    startListening, stopListening, clearLog, addLog,
    checkUpdate, download, cancelDownload, resetState,
  }
})
