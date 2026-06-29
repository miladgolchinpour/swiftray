<template>
  <div class="flex items-center justify-between gap-4 py-1" :class="{ 'opacity-50 pointer-events-none': disabled }">
    <label class="text-sm text-foreground flex-shrink-0">{{ label }}</label>
    <select
      :value="modelValue"
      class="h-8 max-w-[200px] rounded-lg border border-input bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150 cursor-pointer"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option>
    </select>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string | number
  label: string
  options: { label: string; value: string | number }[]
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string | number]
}>()
</script>
