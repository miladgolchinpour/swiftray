<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <!-- Header -->
        <div class="space-y-1">
          <h1 class="text-[28px] font-bold tracking-tight text-foreground">DNS</h1>
          <p class="text-sm text-muted-foreground">Configure DNS resolution settings</p>
        </div>

        <!-- DNS Servers -->
        <SettingsSection title="DNS Servers" icon="network" icon-bg="bg-blue-500">
          <TextFieldRow
            v-model="store.settings.localDNS"
            label="Local DNS"
            placeholder="8.8.8.8"
          />
          <div class="border-t border-border" />
          <TextFieldRow
            v-model="store.settings.remoteDNS"
            label="Remote DNS"
            placeholder="https://dns.google/dns-query"
          />
          <div class="border-t border-border" />
          <TextFieldRow
            v-model="store.settings.bootstrapDNS"
            label="Bootstrap DNS"
            placeholder="https://dns.google/dns-query"
          />
        </SettingsSection>

        <!-- Options -->
        <SettingsSection title="Options" icon="settings" icon-bg="bg-indigo-500">
          <ToggleRow v-model="store.settings.parallelQuery" label="Parallel Query" />
          <div class="border-t border-border" />
          <ToggleRow v-model="store.settings.serveStale" label="Serve Stale" />
        </SettingsSection>

        <!-- Hosts -->
        <SettingsSection title="Hosts" icon="monitor" icon-bg="bg-teal-500">
          <ToggleRow v-model="store.settings.useSystemHosts" label="Use System Hosts" />
          <div class="border-t border-border" />
          <ToggleRow v-model="store.settings.customDNSHosts" label="Custom DNS Hosts" />
        </SettingsSection>

        <!-- FakeIP -->
        <SettingsSection title="FakeIP" icon="wand" icon-bg="bg-pink-500">
          <ToggleRow v-model="store.settings.fakeIP" label="FakeIP" />
          <div class="border-t border-border" />
          <ToggleRow
            v-model="store.settings.blockSVCBHTTPS"
            label="Block SVCB and HTTPS Queries"
            subtitle="Prevent DNS-over-HTTPS lookups"
          />
        </SettingsSection>

        <!-- Regional Domain -->
        <SettingsSection title="Regional Domain" icon="map" icon-bg="bg-amber-600">
          <SelectRow
            :model-value="store.settings.validateRegionalDomain"
            label="Validate Regional Domain IPs"
            :options="regionalOptions"
            @update:model-value="store.settings.validateRegionalDomain = Number($event)"
          />
        </SettingsSection>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import SettingsSection from '@/components/ui/SettingsSection.vue'
import ToggleRow from '@/components/ui/ToggleRow.vue'
import TextFieldRow from '@/components/ui/TextFieldRow.vue'
import SelectRow from '@/components/ui/SelectRow.vue'

const store = useAppStore()

const regionalOptions = [
  { label: 'Empty', value: 0 },
  { label: 'geoip:ir', value: 1 },
]
</script>
