<template>
  <div class="rounded-2xl border border-border bg-card shadow-sm overflow-hidden">
    <div class="px-[18px] py-4 flex items-center gap-2.5">
      <div class="w-7 h-7 rounded-[7px] flex items-center justify-center flex-shrink-0" :class="iconBg">
        <component :is="iconComponent" class="w-3.5 h-3.5 text-white" :stroke-width="2" />
      </div>
      <h3 class="text-sm font-semibold text-card-foreground">{{ title }}</h3>
    </div>
    <div class="border-t border-border" />
    <div class="px-[18px] py-3 space-y-1">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import {
  Hash, Network, Search, Shield, Lock, Puzzle,
  Layout, Radio, Globe, CheckCircle, Zap, Download,
  Info, Link, ExternalLink, Loader2, RefreshCw,
  Sun, Moon, Monitor, X, Check,
} from 'lucide-vue-next'

const iconMap: Record<string, Component> = {
  hash: Hash, network: Network, search: Search, shield: Shield, lock: Lock, puzzle: Puzzle,
  layout: Layout, radio: Radio, globe: Globe, 'check-circle': CheckCircle, zap: Zap,
  download: Download, info: Info, link: Link,
}

const props = defineProps<{
  title: string
  icon: string
  iconBg: string
}>()

const iconComponent = computed(() => iconMap[props.icon] || Info)
</script>
