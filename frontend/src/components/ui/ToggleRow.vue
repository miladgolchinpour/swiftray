<template>
  <div class="flex items-center justify-between py-1" :class="{ 'opacity-50 pointer-events-none': disabled }">
    <div class="space-y-0.5">
      <span class="text-sm text-foreground">{{ label }}</span>
      <p v-if="subtitle" class="text-xs text-muted-foreground">{{ subtitle }}</p>
    </div>
    <button
      class="relative w-10 h-[22px] rounded-full transition-colors duration-150 flex-shrink-0"
      :class="modelValue ? 'bg-blue-600' : 'bg-muted'"
      role="switch"
      :aria-checked="modelValue"
      :disabled="disabled"
      @click="$emit('update:modelValue', !modelValue)"
    >
      <span
        class="absolute top-[3px] left-[3px] w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-150"
        :class="{ 'translate-x-[18px]': modelValue }"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean
  label: string
  subtitle?: string
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script>
