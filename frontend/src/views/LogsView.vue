<template>
  <div class="p-6 max-w-4xl mx-auto space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-lg font-semibold">Logs</h1>
      <button
        class="h-8 px-3 rounded-lg border border-border text-sm font-medium hover:bg-accent transition-colors"
        @click="logsStore.clear()"
      >
        Clear
      </button>
    </div>
    <div class="rounded-xl border border-border bg-card">
      <div v-if="logsStore.entries.length === 0" class="p-6 text-center text-sm text-muted-foreground">
        No logs yet. Start a connection to see xray output.
      </div>
      <div v-else class="divide-y divide-border max-h-[calc(100vh-200px)] overflow-y-auto">
        <div
          v-for="log in logsStore.entries"
          :key="log.id"
          class="px-4 py-2 flex items-start gap-3"
        >
          <span
            class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono uppercase tracking-wide flex-shrink-0 mt-0.5"
            :class="levelClasses[log.level]"
          >
            {{ log.level }}
          </span>
          <span class="text-xs font-mono text-muted-foreground break-all">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLogsStore } from '@/stores/logs'

const logsStore = useLogsStore()

const levelClasses: Record<string, string> = {
  info: 'bg-emerald-500/10 text-emerald-500',
  debug: 'bg-zinc-500/10 text-zinc-400',
  warning: 'bg-amber-500/10 text-amber-500',
  error: 'bg-red-500/10 text-red-500',
}
</script>
