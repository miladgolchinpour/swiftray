import { defineStore } from 'pinia'
import { ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { api } from '@/api/client'

export interface LogEntry {
  id: string
  level: 'info' | 'debug' | 'warning' | 'error'
  message: string
  timestamp: string
}

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<LogEntry[]>([])
  let listening = false

  function startListening() {
    if (listening) return
    listening = true

    EventsOn('log:entry', (dataStr: string) => {
      try {
        const entry: LogEntry = JSON.parse(dataStr)
        entries.value.unshift(entry)
        if (entries.value.length > 200) {
          entries.value = entries.value.slice(0, 200)
        }
      } catch (e) {
        console.error('[Logs] Failed to parse log entry:', e)
      }
    })
  }

  async function loadInitial() {
    const { data } = await api.getLogs()
    if (data) {
      entries.value = data as unknown as LogEntry[]
    }
  }

  async function clear() {
    entries.value = []
    await api.clearLogs()
  }

  return {
    entries,
    startListening,
    loadInitial,
    clear,
  }
})
