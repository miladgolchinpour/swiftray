<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <!-- Header -->
        <div class="space-y-1">
          <h1 class="text-[28px] font-bold tracking-tight text-foreground">Routing</h1>
          <p class="text-sm text-muted-foreground">Configure how traffic is routed through proxy nodes</p>
        </div>

        <!-- Rule Section -->
        <SettingsSection title="Rule" icon="route" icon-bg="bg-orange-500">
          <div class="flex items-center justify-between py-1">
            <span class="text-sm text-foreground">Routing Mode</span>
            <div class="inline-flex gap-2">
              <button
                v-for="mode in routingModeOptions"
                :key="mode.value"
                class="h-8 px-3 rounded-lg text-xs font-medium transition-all duration-150 inline-flex items-center gap-1.5"
                :class="store.settings.routingMode === mode.value
                  ? 'bg-accent text-accent-foreground shadow-sm border border-border'
                  : 'text-muted-foreground hover:text-foreground border border-transparent'"
                @click="store.settings.routingMode = mode.value"
              >
                <component :is="mode.icon" class="w-3.5 h-3.5" />
                {{ mode.label }}
              </button>
            </div>
          </div>
          <div class="border-t border-border" />
          <SelectRow
            :model-value="store.settings.domainStrategy"
            label="Domain Strategy"
            :options="domainStrategyOptions"
            @update:model-value="store.settings.domainStrategy = Number($event)"
          />
        </SettingsSection>

        <!-- Exclusions (Rule-Based only) -->
        <template v-if="store.settings.routingMode === 1">
          <SettingsSection title="Exclusions" icon="shield-x" icon-bg="bg-red-500">
            <div class="space-y-2">
              <p class="text-xs text-muted-foreground">Addresses to bypass the proxy, one per line.</p>
              <textarea
                v-model="store.settings.exclusions"
                rows="6"
                class="w-full rounded-xl border border-input bg-background p-3 text-xs font-mono text-foreground placeholder:text-muted-foreground resize-none focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
                style="min-height: 120px;"
                placeholder="localhost&#10;127.0.0.0/8&#10;::1"
              />
            </div>
          </SettingsSection>

          <!-- Bypass Toggles -->
          <SettingsSection title="Bypass Toggles" icon="shield-check" icon-bg="bg-green-500">
            <ToggleRow v-model="store.settings.bypassIran" label="Bypass Iranian websites" subtitle="Routes .ir domains directly" />
            <div class="border-t border-border" />
            <ToggleRow v-model="store.settings.bypassRussia" label="Bypass Russian websites" subtitle="Routes .ru domains directly" />
            <div class="border-t border-border" />
            <ToggleRow v-model="store.settings.bypassChina" label="Bypass Chinese websites" subtitle="Routes .cn domains directly" />
          </SettingsSection>
        </template>

        <!-- Rules Display -->
        <SettingsSection title="Rules" icon="list" icon-bg="bg-purple-500">
          <!-- Global Mode Info -->
          <div
            v-if="store.settings.routingMode === 0"
            class="flex flex-col items-center py-8 gap-4 rounded-xl bg-muted/30"
          >
            <Globe class="w-9 h-9 text-muted-foreground/40" />
            <div class="text-center space-y-1">
              <p class="text-sm font-semibold text-foreground">Global Mode</p>
              <p class="text-xs text-muted-foreground">All traffic is routed through the proxy node</p>
            </div>
          </div>

          <!-- Rule-Based Rules List -->
          <div v-else class="space-y-3">
            <p class="text-xs font-medium text-muted-foreground">Active Bypass Rules</p>
            <div class="rounded-xl bg-muted/30 p-3.5 space-y-2.5">
              <!-- Always present: private IPs -->
              <div class="flex items-center gap-2">
                <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                <span class="text-xs font-mono text-foreground">geoip:private</span>
              </div>

              <!-- Custom exclusions -->
              <div v-if="trimmedExclusions.length > 0" class="flex items-center gap-2">
                <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                <span class="text-xs text-foreground">Custom exclusions from above</span>
              </div>

              <!-- Iran bypass -->
              <template v-if="store.settings.bypassIran">
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geosite:geolocation-ir + domain:.ir</span>
                </div>
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geoip:ir</span>
                </div>
              </template>

              <!-- Russia bypass -->
              <template v-if="store.settings.bypassRussia">
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geosite:geolocation-ru + domain:.ru</span>
                </div>
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geoip:ru</span>
                </div>
              </template>

              <!-- China bypass -->
              <template v-if="store.settings.bypassChina">
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geosite:geolocation-cn + domain:.cn</span>
                </div>
                <div class="flex items-center gap-2">
                  <CheckCircle class="w-3.5 h-3.5 text-green-500 flex-shrink-0" />
                  <span class="text-xs font-mono text-foreground">geoip:cn</span>
                </div>
              </template>

              <!-- Catch-all -->
              <div class="flex items-center gap-2 pt-1 border-t border-border/50">
                <ArrowRight class="w-3.5 h-3.5 text-blue-500 flex-shrink-0" />
                <span class="text-xs text-muted-foreground">All other traffic is routed through the proxy node.</span>
              </div>
            </div>
          </div>
        </SettingsSection>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'
import SettingsSection from '@/components/ui/SettingsSection.vue'
import ToggleRow from '@/components/ui/ToggleRow.vue'
import SelectRow from '@/components/ui/SelectRow.vue'
import { Globe, List, CheckCircle, ArrowRight } from 'lucide-vue-next'

const store = useAppStore()

const routingModeOptions = [
  { label: 'Global', value: 0, icon: Globe },
  { label: 'Rule-Based', value: 1 },
]

const domainStrategyOptions = [
  { label: 'AsIs', value: 0 },
  { label: 'IPIfNonMatch', value: 1 },
  { label: 'IPOnDemand', value: 2 },
]

const trimmedExclusions = computed(() =>
  store.settings.exclusions
    .split('\n')
    .map(l => l.trim())
    .filter(Boolean)
)
</script>
