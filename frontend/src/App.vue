<template>
  <div class="flex h-screen overflow-hidden">
    <NotificationToast />
    <Sidebar />
    <main class="flex-1 overflow-y-auto">
      <div v-if="isLoading" class="flex items-center justify-center h-full">
        <div class="text-center space-y-3">
          <div class="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin mx-auto" />
          <p class="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
      <router-view v-else />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import Sidebar from '@/components/layout/Sidebar.vue'
import NotificationToast from '@/components/ui/NotificationToast.vue'
import { useStartup } from '@/composables/useStartup'
import { useAppStore } from '@/stores/app'
import { useConnectionStore } from '@/stores/connection'
import { useLogsStore } from '@/stores/logs'
import { useIPInfoStore } from '@/stores/ipinfo'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

const { isLoading, initialize } = useStartup()
const appStore = useAppStore()
const connectionStore = useConnectionStore()
const logsStore = useLogsStore()
const ipInfoStore = useIPInfoStore()

// --- Theme ---
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

function applyTheme() {
  const saved = localStorage.getItem('swiftray-theme') as 'light' | 'dark' | 'system' | null
  const mode = saved || 'dark'
  const root = document.documentElement
  if (mode === 'light') {
    root.classList.remove('dark')
  } else if (mode === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.toggle('dark', mediaQuery.matches)
  }
}

function onSystemThemeChange() {
  const saved = localStorage.getItem('swiftray-theme')
  if (!saved || saved === 'system') {
    applyTheme()
  }
}

onMounted(async () => {
  // Apply theme immediately
  applyTheme()
  mediaQuery.addEventListener('change', onSystemThemeChange)

  connectionStore.startListening()
  logsStore.startListening()
  ipInfoStore.startListening()

  // Listen for proxy toggle from tray menu bar
  EventsOn('settings:proxy-toggled', (dataStr: string) => {
    try {
      const data = JSON.parse(dataStr)
      appStore.settings.enableSystemProxy = data.enableSystemProxy
    } catch (e) {
      console.error('[App] Failed to parse proxy toggle event:', e)
    }
  })

  await initialize()
  if (!isLoading.value) {
    await Promise.all([
      appStore.loadAll(),
      connectionStore.syncState(),
      logsStore.loadInitial(),
      ipInfoStore.loadCached(),
    ])
  }
})

onUnmounted(() => {
  mediaQuery.removeEventListener('change', onSystemThemeChange)
  EventsOff('settings:proxy-toggled')
})
</script>
