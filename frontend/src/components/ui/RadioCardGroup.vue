<template>
  <div class="inline-flex gap-1 p-0.5 rounded-lg bg-muted">
    <button
      v-for="option in options"
      :key="String(option.value)"
      class="h-8 px-3 rounded-md text-xs font-medium transition-all duration-150 inline-flex items-center gap-1.5"
      :class="modelValue === option.value
        ? 'bg-card text-card-foreground shadow-sm'
        : 'text-muted-foreground hover:text-foreground'"
      @click="$emit('update:modelValue', option.value)"
    >
      <component v-if="option.icon" :is="option.icon" class="w-3.5 h-3.5" />
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  modelValue: string | number
  options: { label: string; value: string | number; icon?: Component }[]
}>()

defineEmits<{
  'update:modelValue': [value: string | number]
}>()
</script>
