<template>
  <div
    class="rounded-2xl border border-border bg-card shadow-sm overflow-hidden transition-shadow duration-200 hover:shadow-md"
  >
    <div
      class="flex items-center gap-2.5 px-[18px] py-[18px] cursor-pointer select-none"
      role="button"
      :tabindex="0"
      :aria-expanded="!collapsed"
      @click="$emit('update:collapsed', !collapsed)"
      @keydown.enter="$emit('update:collapsed', !collapsed)"
      @keydown.space.prevent="$emit('update:collapsed', !collapsed)"
    >
      <div
        v-if="icon"
        class="w-7 h-7 rounded-[7px] flex items-center justify-center flex-shrink-0"
        :class="iconBgClass"
      >
        <component :is="icon" class="w-3.5 h-3.5 text-white" :stroke-width="2" />
      </div>

      <h3 class="text-sm font-semibold text-card-foreground">{{ title }}</h3>
      <span v-if="count !== undefined" class="text-xs text-muted-foreground">
        ({{ count }})
      </span>

      <div class="flex-1" />

      <slot name="actions" />

      <div
        class="w-5 h-5 flex items-center justify-center text-muted-foreground transition-transform duration-150"
        :class="{ '-rotate-90': collapsed }"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
        </svg>
      </div>
    </div>

    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="opacity-0 -translate-y-1"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div v-show="!collapsed">
        <div class="border-t border-border" />
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  icon?: Component
  iconBgClass?: string
  count?: number
  collapsed: boolean
}>()

defineEmits<{
  'update:collapsed': [value: boolean]
}>()
</script>
