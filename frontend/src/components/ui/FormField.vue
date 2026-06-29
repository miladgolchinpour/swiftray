<template>
  <div class="space-y-1.5">
    <label v-if="label" class="text-xs font-medium text-muted-foreground">{{ label }}</label>
    <slot />
    <p v-if="error" class="text-xs text-red-500">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label?: string
  error?: string
}>()
</script>
