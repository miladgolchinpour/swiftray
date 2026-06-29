import { defineStore } from 'pinia'
import { ref } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { api } from '@/api/client'
import type { IPInfo } from '@/types'

export const useIPInfoStore = defineStore('ipinfo', () => {
  const info = ref<IPInfo | null>(null)
  const loading = ref(false)
  let listening = false

  function startListening() {
    if (listening) return
    listening = true

    EventsOn('ipinfo:update', (dataStr: string) => {
      try {
        const data = JSON.parse(dataStr)
        info.value = data
      } catch (e) {
        console.error('[IPInfo] Failed to parse:', e)
      }
    })
  }

  async function fetch() {
    loading.value = true
    const { data } = await api.fetchIPInfo()
    if (data) {
      info.value = data
    }
    loading.value = false
  }

  async function loadCached() {
    const { data } = await api.getIPInfo()
    if (data) {
      info.value = data
    }
  }

  return {
    info,
    loading,
    startListening,
    fetch,
    loadCached,
  }
})
