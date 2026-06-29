<template>
  <router-link
    :to="path"
    class="flex items-center gap-2.5 px-3 py-1.5 rounded-md text-sm transition-colors"
    :class="isActive
      ? 'bg-accent text-accent-foreground font-medium'
      : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'"
  >
    <component :is="icon" class="w-4 h-4 flex-shrink-0" />
    <span>{{ label }}</span>
  </router-link>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { Component } from 'vue'

const props = defineProps<{
  icon: Component
  label: string
  path: string
}>()

const route = useRoute()
const isActive = computed(() => route.path === props.path)
</script>
