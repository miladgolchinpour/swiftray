<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm">
      <TransitionGroup
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 translate-x-4"
        enter-to-class="opacity-100 translate-x-0"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 translate-x-0"
        leave-to-class="opacity-0 translate-x-4"
      >
        <div
          v-for="n in notifications"
          :key="n.id"
          class="rounded-lg border border-border bg-card shadow-lg p-3 flex items-start gap-2.5 cursor-pointer"
          @click="remove(n.id)"
        >
          <div class="mt-0.5">
            <CheckCircle2 v-if="n.type === 'success'" class="w-4 h-4 text-emerald-500" />
            <XCircle v-else-if="n.type === 'error'" class="w-4 h-4 text-red-500" />
            <AlertTriangle v-else-if="n.type === 'warning'" class="w-4 h-4 text-amber-500" />
            <Info v-else class="w-4 h-4 text-blue-500" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium">{{ n.title }}</p>
            <p v-if="n.message" class="text-xs text-muted-foreground mt-0.5">{{ n.message }}</p>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { CheckCircle2, XCircle, AlertTriangle, Info } from 'lucide-vue-next'
import { useNotifications } from '@/composables/useNotifications'

const { notifications, remove } = useNotifications()
</script>
