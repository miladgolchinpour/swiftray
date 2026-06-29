<template>
  <div class="p-3.5 rounded-xl border border-border bg-card transition-all duration-200">
    <div class="flex items-center gap-2 mb-1.5">
      <Radio class="w-3.5 h-3.5 text-muted-foreground animate-pulse" />
      <span class="text-xs text-muted-foreground">Testing nodes...</span>
      <span class="text-xs text-muted-foreground ml-auto tabular-nums">{{ tested }}/{{ total }}</span>
      <span class="text-xs text-muted-foreground tabular-nums">{{ percent }}%</span>
      <button
        class="w-5 h-5 rounded flex items-center justify-center text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors duration-150"
        title="Cancel test"
        @click="$emit('cancel')"
      >
        <X class="w-3 h-3" />
      </button>
    </div>
    <div class="h-2 rounded bg-muted overflow-hidden">
      <div
        class="h-full rounded bg-gradient-to-r from-blue-500 to-purple-500 transition-all duration-300 ease-out"
        :style="{ width: `${percent}%` }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Radio, X } from 'lucide-vue-next'

const props = defineProps<{
  tested: number
  total: number
}>()

defineEmits<{
  cancel: []
}>()

const percent = computed(() =>
  props.total > 0 ? Math.round((props.tested / props.total) * 100) : 0
)
</script>
