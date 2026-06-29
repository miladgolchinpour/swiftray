<template>
  <div class="flex items-center justify-between gap-4 py-1" :class="{ 'opacity-50 pointer-events-none': disabled }">
    <label class="text-sm text-foreground flex-shrink-0">{{ label }}</label>
    <div class="inline-flex items-center gap-0">
      <button
        class="w-7 h-7 rounded-l-lg border border-border bg-card flex items-center justify-center text-xs font-semibold text-card-foreground hover:bg-accent transition-colors duration-150 disabled:opacity-40 disabled:cursor-not-allowed"
        :disabled="modelValue <= min"
        @click="$emit('update:modelValue', modelValue - step)"
      >
        −
      </button>
      <div class="w-12 h-7 border-y border-border bg-background flex items-center justify-center text-xs font-mono text-foreground tabular-nums">
        {{ displayValue }}
      </div>
      <button
        class="w-7 h-7 rounded-r-lg border border-border bg-card flex items-center justify-center text-xs font-semibold text-card-foreground hover:bg-accent transition-colors duration-150 disabled:opacity-40 disabled:cursor-not-allowed"
        :disabled="modelValue >= max"
        @click="$emit('update:modelValue', modelValue + step)"
      >
        +
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: number
  label: string
  min: number
  max: number
  step: number
  format?: string
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: number]
}>()

const displayValue = computed(() => {
  if (props.format === 'decimal') {
    return props.modelValue.toFixed(1)
  }
  return String(props.modelValue)
})
</script>
