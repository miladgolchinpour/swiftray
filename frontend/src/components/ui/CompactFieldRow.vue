<template>
  <div class="flex items-center justify-between gap-4 py-1" :class="{ 'opacity-50 pointer-events-none': disabled }">
    <label class="text-sm text-foreground flex-shrink-0">{{ label }}</label>
    <input
      :value="modelValue"
      type="text"
      :placeholder="placeholder"
      class="h-8 w-20 rounded-lg border border-input bg-background px-3 text-sm text-right text-foreground font-mono placeholder:text-muted-foreground placeholder:font-sans focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string
  label: string
  placeholder?: string
  disabled?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
