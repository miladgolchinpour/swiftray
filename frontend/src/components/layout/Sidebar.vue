<template>
  <aside class="w-56 flex-shrink-0 border-r border-border bg-card flex flex-col">
    <div class="px-4 py-4 border-b border-border">
      <div class="flex items-center gap-2.5">
        <img src="@/assets/images/logo.png" alt="SwiftRay" class="w-7 h-7 rounded-lg object-contain" />
        <span class="text-sm font-semibold tracking-tight">SwiftRay</span>
      </div>
    </div>

    <nav class="flex-1 px-2 py-2 space-y-0.5 overflow-y-auto">
      <SidebarItem
        v-for="item in navItems"
        :key="item.path"
        :icon="item.icon"
        :label="item.label"
        :path="item.path"
      />
    </nav>

    <div class="px-3 py-3 border-t border-border">
      <div class="flex items-center gap-2 px-2">
        <div
          class="w-2 h-2 rounded-full"
          :class="stateClasses"
        />
        <span class="text-xs text-muted-foreground">
          {{ stateLabel }}
        </span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { LayoutDashboard, Rss, Server, Route, Wifi, ScrollText, Settings } from 'lucide-vue-next'
import SidebarItem from './SidebarItem.vue'
import { useConnectionStore } from '@/stores/connection'

const connectionStore = useConnectionStore()

const stateClasses = computed(() => {
  switch (connectionStore.state) {
    case 'connected': return 'bg-emerald-500'
    case 'connecting':
    case 'disconnecting': return 'bg-amber-500 animate-pulse'
    case 'error': return 'bg-red-500'
    default: return 'bg-zinc-400'
  }
})

const stateLabel = computed(() => {
  switch (connectionStore.state) {
    case 'connected': return 'Connected'
    case 'connecting': return 'Connecting...'
    case 'disconnecting': return 'Disconnecting...'
    case 'error': return 'Error'
    default: return 'Disconnected'
  }
})

const navItems = [
  { icon: LayoutDashboard, label: 'Dashboard', path: '/dashboard' },
  { icon: Rss, label: 'Subscriptions', path: '/subscriptions' },
  { icon: Server, label: 'Local Nodes', path: '/nodes' },
  { icon: Route, label: 'Routing', path: '/routing' },
  { icon: Wifi, label: 'DNS', path: '/dns' },
  { icon: ScrollText, label: 'Logs', path: '/logs' },
  { icon: Settings, label: 'Settings', path: '/settings' },
]
</script>
