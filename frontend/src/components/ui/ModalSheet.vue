<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @keydown.esc="$emit('update:modelValue', false)"
      >
        <div
          class="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity duration-150"
          @click="$emit('update:modelValue', false)"
        />
        <div
          ref="panelRef"
          class="relative z-10 bg-card border border-border rounded-2xl shadow-lg flex flex-col overflow-hidden transition-all duration-150 max-h-[85vh]"
          :style="{ width: `${width}px` }"
          tabindex="-1"
        >
          <div class="flex flex-col items-center pt-7 pb-5 px-6">
            <slot name="icon">
              <div
                v-if="icon"
                class="w-14 h-14 rounded-2xl flex items-center justify-center mb-3"
                :class="iconBgClass"
              >
                <component :is="icon" class="w-7 h-7 text-white" :stroke-width="1.5" />
              </div>
            </slot>
            <slot name="header">
              <h2 v-if="title" class="text-lg font-bold text-card-foreground">{{ title }}</h2>
            </slot>
          </div>

          <div class="border-t border-border" />

          <div class="flex-1 overflow-y-auto overflow-x-auto px-6 py-5">
            <slot />
          </div>

          <div class="border-t border-border" />

          <div class="flex items-center justify-between px-6 py-4">
            <slot name="footer-left">
              <button
                class="h-9 px-4 rounded-lg text-sm font-medium border border-border bg-card text-card-foreground hover:bg-accent hover:text-accent-foreground transition-colors duration-150"
                @click="$emit('update:modelValue', false)"
              >
                {{ cancelLabel }}
              </button>
            </slot>
            <slot name="footer-right">
              <button
                :disabled="disabled || loading"
                class="h-9 px-5 rounded-lg text-sm font-semibold text-white transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
                :class="actionBtnClass"
                @click="$emit('action')"
              >
                <span v-if="loading" class="inline-flex items-center gap-2">
                  <span class="w-3.5 h-3.5 border-2 border-white/40 border-t-white rounded-full animate-spin" />
                  {{ loadingLabel }}
                </span>
                <span v-else>{{ actionLabel }}</span>
              </button>
            </slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, type Component } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  icon?: Component
  iconBgClass?: string
  width?: number
  actionLabel?: string
  actionBtnClass?: string
  cancelLabel?: string
  loadingLabel?: string
  disabled?: boolean
  loading?: boolean
}>(), {
  width: 460,
  actionLabel: 'Save',
  cancelLabel: 'Cancel',
  loadingLabel: 'Saving...',
  iconBgClass: 'bg-blue-500',
})

defineEmits<{
  'update:modelValue': [value: boolean]
  'action': []
}>()

const panelRef = ref<HTMLElement>()

watch(() => props.modelValue, async (open) => {
  if (open) {
    await nextTick()
    panelRef.value?.focus()
  }
})
</script>


