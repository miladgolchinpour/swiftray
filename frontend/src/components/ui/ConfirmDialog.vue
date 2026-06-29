<template>
  <ModalSheet
    :model-value="modelValue"
    :width="400"
    icon-bg-class="bg-red-500"
    :action-label="confirmLabel"
    action-btn-class="bg-red-600 hover:bg-red-700"
    cancel-label="Cancel"
    :loading="loading"
    :loading-label="loading ? 'Deleting...' : ''"
    @update:model-value="$emit('update:modelValue', $event)"
    @action="$emit('confirm')"
  >
    <template #icon>
      <div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3 bg-red-500">
        <svg class="w-7 h-7 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
        </svg>
      </div>
    </template>
    <template #header>
      <h2 class="text-lg font-bold text-card-foreground">{{ title }}</h2>
    </template>
    <p class="text-sm text-muted-foreground text-center">{{ message }}</p>
  </ModalSheet>
</template>

<script setup lang="ts">
import ModalSheet from './ModalSheet.vue'

defineProps<{
  modelValue: boolean
  title: string
  message: string
  confirmLabel?: string
  loading?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': []
}>()
</script>
