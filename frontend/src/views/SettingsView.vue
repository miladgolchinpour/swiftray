<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <div class="flex items-start justify-between">
          <div class="space-y-1">
            <h1 class="text-[28px] font-bold tracking-tight text-foreground">Settings</h1>
            <p class="text-sm text-muted-foreground">Configure application and proxy behavior</p>
          </div>
          <div class="grid grid-cols-4 gap-1 p-0.5 rounded-lg bg-muted">
            <button v-for="(tab, i) in tabs" :key="tab"
              class="h-8 px-3 rounded-md text-xs font-medium transition-all duration-150"
              :class="selectedTab === i ? 'bg-card text-card-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
              @click="selectedTab = i">{{ tab }}</button>
          </div>
        </div>

        <!-- BASIC TAB -->
        <template v-if="selectedTab === 0">
          <SettingsSection title="Port" icon="hash" icon-bg="bg-blue-500">
            <ToggleRow v-model="store.settings.mixedPort" label="Mixed Port (HTTP + SOCKS)" />
            <div class="border-t border-border" />
            <template v-if="store.settings.mixedPort">
              <CompactFieldRow :model-value="String(store.settings.socksPort)" label="Port" placeholder="1080" @update:model-value="store.settings.socksPort = parseInt($event) || 1080" />
            </template>
            <template v-else>
              <CompactFieldRow :model-value="String(store.settings.socksPort)" label="SOCKS Port" placeholder="1080" @update:model-value="store.settings.socksPort = parseInt($event) || 1080" />
              <div class="border-t border-border" />
              <CompactFieldRow :model-value="String(store.settings.httpPort)" label="HTTP Port" placeholder="1081" @update:model-value="store.settings.httpPort = parseInt($event) || 1081" />
            </template>
          </SettingsSection>
          <SettingsSection title="Network" icon="network" icon-bg="bg-teal-500">
            <ToggleRow v-model="store.settings.enableUDP" label="Enable UDP" />
            <div class="border-t border-border" />
            <ToggleRow v-model="store.settings.allowLAN" label="Allow LAN" subtitle="Binds to 0.0.0.0" />
            <div class="border-t border-border" />
            <ToggleRow v-model="store.settings.routeOnly" label="Route Only" />
          </SettingsSection>
          <SettingsSection title="Sniffing" icon="search" icon-bg="bg-indigo-500">
            <ToggleRow v-model="store.settings.enableSniffing" label="Enable Sniffing" />
            <div class="border-t border-border" />
            <div class="pl-3 space-y-1" :class="{ 'opacity-50 pointer-events-none': !store.settings.enableSniffing }">
              <ToggleRow v-model="store.settings.sniffHTTP" label="HTTP" />
              <ToggleRow v-model="store.settings.sniffTLS" label="TLS" />
              <ToggleRow v-model="store.settings.sniffQUIC" label="QUIC" />
              <ToggleRow v-model="store.settings.sniffFakeDNS" label="FakeDNS" />
            </div>
          </SettingsSection>
          <SettingsSection title="Authentication" icon="shield" icon-bg="bg-orange-500">
            <ToggleRow v-model="store.settings.useProxyAuth" label="Use Proxy Auth" />
            <div class="border-t border-border" />
            <div class="pl-3 space-y-1" :class="{ 'opacity-50 pointer-events-none': !store.settings.useProxyAuth }">
              <TextFieldRow v-model="store.settings.proxyUsername" label="Username" placeholder="" />
              <div class="border-t border-border" />
              <TextFieldRow v-model="store.settings.proxyPassword" label="Password" placeholder="" type="password" />
            </div>
          </SettingsSection>
          <SettingsSection title="TLS" icon="lock" icon-bg="bg-green-500">
            <SelectRow :model-value="store.settings.defaultFingerprint" label="Default Fingerprint" :options="fingerprintOptions" @update:model-value="store.settings.defaultFingerprint = String($event)" />
          </SettingsSection>
          <SettingsSection title="Fragment" icon="puzzle" icon-bg="bg-purple-500">
            <ToggleRow v-model="store.settings.enableFragment" label="Enable Fragment" />
            <div class="border-t border-border" />
            <div class="pl-3 space-y-1" :class="{ 'opacity-50 pointer-events-none': !store.settings.enableFragment }">
              <CompactFieldRow v-model="store.settings.fragmentPackLength" label="Length" placeholder="100-200" />
              <div class="border-t border-border" />
              <CompactFieldRow v-model="store.settings.fragmentSleep" label="Sleep" placeholder="50-100" />
              <div class="border-t border-border" />
              <CompactFieldRow v-model="store.settings.fragmentInterval" label="Interval" placeholder="1-5" />
            </div>
          </SettingsSection>
        </template>

        <!-- CLIENT TAB -->
        <template v-if="selectedTab === 1">
          <SettingsSection title="Interface" icon="layout" icon-bg="bg-cyan-500">
            <div class="flex items-center justify-between py-1">
              <span class="text-sm text-foreground">Theme</span>
              <div class="grid grid-cols-3 gap-1 p-0.5 rounded-lg bg-muted">
                <button v-for="mode in themeOptions" :key="mode.value"
                  class="h-7 px-2.5 rounded-md text-xs font-medium transition-all duration-150 inline-flex items-center gap-1"
                  :class="currentTheme === mode.value ? 'bg-card text-card-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                  @click="setTheme(mode.value)">
                  <component :is="mode.icon" class="w-3 h-3" />{{ mode.label }}
                </button>
              </div>
            </div>
            <div class="border-t border-border" />
            <ToggleRow v-if="platformDisplay !== 'Linux'" v-model="store.settings.enableMenuBar" label="Enable Menu Bar" subtitle="Show status in menu bar" />
          </SettingsSection>
          <SettingsSection title="Testing" icon="radio" icon-bg="bg-yellow-500">
            <SelectRow :model-value="store.settings.urlTestMode" label="URL Test Mode" :options="urlTestModeOptions" @update:model-value="store.settings.urlTestMode = String($event)" />
            <div class="border-t border-border" />
            <TextFieldRow v-model="store.settings.pingTestURL" label="Ping Test URL" placeholder="https://www.google.com/generate_204" />
            <div class="border-t border-border" />
            <NumberRow :model-value="store.settings.urlTestTimeout" label="URL Test Timeout" :min="1" :max="10" :step="0.5" format="decimal" @update:model-value="store.settings.urlTestTimeout = $event" />
            <div class="border-t border-border" />
            <NumberRow :model-value="store.settings.urlTestConcurrency" label="Concurrent Tests" :min="1" :max="50" :step="1" @update:model-value="store.settings.urlTestConcurrency = $event" />
          </SettingsSection>
          <SettingsSection title="Geo Files" icon="globe" icon-bg="bg-green-500">
            <TextFieldRow v-model="store.settings.customGeoSources" label="Custom Geo Sources" placeholder="Optional" />
          </SettingsSection>
        </template>

        <!-- XRAY TAB -->
        <template v-if="selectedTab === 2">
          <!-- Status -->
          <SettingsSection title="Status" icon="check-circle" icon-bg="bg-green-500">
            <div class="flex items-center gap-2.5 py-1">
              <div class="w-5 h-5 rounded-full flex items-center justify-center" :class="resourceReady ? 'bg-green-500/15' : 'bg-red-500/15'">
                <Check v-if="resourceReady" class="w-3 h-3 text-green-500" />
                <X v-else class="w-3 h-3 text-red-500" />
              </div>
              <span class="text-sm" :class="resourceReady ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'">
                {{ resourceReady ? 'Bundled resources are ready' : 'Bundled resources missing' }}
              </span>
            </div>
          </SettingsSection>

          <!-- Runtime -->
          <SettingsSection title="Runtime" icon="zap" icon-bg="bg-orange-500">
            <div class="flex items-center justify-between py-1">
              <div>
                <p class="text-sm font-medium text-foreground">Xray Runtime</p>
                <p class="text-xs text-muted-foreground">{{ xrayInstalled ? `v${xrayVersionDisplay || 'unknown'}` : 'Not installed' }}</p>
              </div>
              <CheckCircle v-if="xrayInstalled" class="w-5 h-5 text-green-500" />
              <XCircle v-else class="w-5 h-5 text-red-500" />
            </div>
            <div class="border-t border-border" />
            <div class="flex gap-5 py-1">
              <div><p class="text-xs font-mono text-foreground">xray</p><p class="text-[10px]" :class="resourceStatus?.xrayExists ? 'text-green-500' : 'text-red-500'">{{ resourceStatus?.xrayExists ? 'Present' : 'Missing' }}</p></div>
              <div><p class="text-xs font-mono text-foreground">geoip.dat</p><p class="text-[10px]" :class="resourceStatus?.geoIPExists ? 'text-green-500' : 'text-red-500'">{{ resourceStatus?.geoIPExists ? 'Present' : 'Missing' }}</p></div>
              <div><p class="text-xs font-mono text-foreground">geosite.dat</p><p class="text-[10px]" :class="resourceStatus?.geoSiteExists ? 'text-green-500' : 'text-red-500'">{{ resourceStatus?.geoSiteExists ? 'Present' : 'Missing' }}</p></div>
            </div>
            <div class="border-t border-border" />
            <UpdaterSection />
          </SettingsSection>
        </template>

        <!-- ABOUT TAB -->
        <template v-if="selectedTab === 3">
          <div class="flex items-center gap-3.5">
            <img src="@/assets/images/logo.png" alt="SwiftRay" class="w-16 h-16 rounded-2xl object-contain" />
            <div class="space-y-1"><h2 class="text-lg font-bold text-foreground">SwiftRay</h2><p class="text-sm text-muted-foreground">A light Xray client.</p></div>
          </div>
          <SettingsSection title="Application" icon="info" icon-bg="bg-blue-500">
            <div class="flex items-center justify-between py-1"><span class="text-sm text-foreground">Version</span><span class="text-sm text-muted-foreground">{{ appVersion }}</span></div>
            <div class="border-t border-border" />
            <div class="flex items-center justify-between py-1"><span class="text-sm text-foreground">Platform</span><span class="text-sm text-muted-foreground">{{ platformDisplay }}</span></div>
            <div class="border-t border-border" />
            <button :disabled="checkingAppUpdate" class="w-full h-10 rounded-lg bg-blue-600 text-sm font-semibold text-white hover:bg-blue-700 transition-all duration-150 disabled:opacity-50 inline-flex items-center justify-center gap-2" @click="checkAppUpdate">
              <Loader2 v-if="checkingAppUpdate" class="w-4 h-4 animate-spin" /><RefreshCw v-else class="w-4 h-4" />
              {{ checkingAppUpdate ? 'Checking...' : 'Check for Updates' }}
            </button>
            <p v-if="appUpdateMsg" class="text-xs text-muted-foreground flex items-center gap-1.5 mt-1">
              <CheckCircle v-if="appUpdateMsg.includes('up to date')" class="w-3.5 h-3.5 text-green-500" /><Download v-else class="w-3.5 h-3.5 text-blue-500" />{{ appUpdateMsg }}
            </p>
          </SettingsSection>
          <SettingsSection title="Links" icon="link" icon-bg="bg-green-500">
            <button class="w-full h-10 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent transition-all duration-150 inline-flex items-center justify-center gap-2" @click="openGitHub">
              <ExternalLink class="w-4 h-4" />GitHub Repository
            </button>
          </SettingsSection>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useConnectionStore } from '@/stores/connection'
import { useUpdaterStore } from '@/stores/updater'
import { useNotifications } from '@/composables/useNotifications'
import { api } from '@/api/client'
import SettingsSection from '@/components/ui/SettingsSection.vue'
import ToggleRow from '@/components/ui/ToggleRow.vue'
import TextFieldRow from '@/components/ui/TextFieldRow.vue'
import SelectRow from '@/components/ui/SelectRow.vue'
import NumberRow from '@/components/ui/NumberRow.vue'
import CompactFieldRow from '@/components/ui/CompactFieldRow.vue'
import UpdaterSection from '@/components/ui/UpdaterSection.vue'
import { Check, X, CheckCircle, XCircle, Loader2, RefreshCw, ExternalLink, Download, Sun, Moon, Monitor } from 'lucide-vue-next'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

const store = useAppStore()
const connectionStore = useConnectionStore()
const updaterStore = useUpdaterStore()
const notify = useNotifications()

const tabs = ['Basic', 'Client', 'Xray', 'About']
const selectedTab = ref(0)
const fingerprintOptions = [
  { label: 'chrome', value: 'chrome' }, { label: 'firefox', value: 'firefox' },
  { label: 'safari', value: 'safari' }, { label: 'edge', value: 'edge' },
  { label: 'random', value: 'random' }, { label: 'randomized', value: 'randomized' },
]
const urlTestModeOptions = [
  { label: 'TCP Connect', value: 'tcp' },
  { label: 'HTTP GET', value: 'http' },
]

const resourceReady = computed(() => resourceStatus.value?.xrayExists && resourceStatus.value?.geoIPExists && resourceStatus.value?.geoSiteExists)
const xrayInstalled = computed(() => resourceStatus.value?.xrayExists ?? false)
const resourceStatus = ref<any>(null)
const xrayVersionDisplay = ref('')
const appVersion = ref('0.1.0')
const platformDisplay = computed(() => { const ua = navigator.userAgent; if (ua.includes('Mac')) return 'macOS'; if (ua.includes('Win')) return 'Windows'; if (ua.includes('Linux')) return 'Linux'; return 'Unknown' })

onMounted(async () => {
  updaterStore.startListening()
  const { data } = await api.getResourceStatus(); resourceStatus.value = data
  const info = await api.getAppInfo(); if (info.data) xrayVersionDisplay.value = (info.data as any).xrayVersion || ''
})
onUnmounted(() => { updaterStore.stopListening() })

// --- App Update ---
const checkingAppUpdate = ref(false)
const appUpdateMsg = ref('')

function compareVersions(a: String, b: String) {
  const aParts = a.split('.').map(Number)
  const bParts = b.split('.').map(Number)
  const length = Math.max(aParts.length, bParts.length)

  for (let i = 0; i < length; i++) {
    const x = aParts[i] ?? 0
    const y = bParts[i] ?? 0

    if (x < y) return -1
    if (x > y) return 1
  }

  return 0
}

async function checkAppUpdate() {
  checkingAppUpdate.value = true
  appUpdateMsg.value = ''

  try {
    const r = await fetch(
      'https://api.github.com/repos/MiladGolchinpour/swiftray/releases/latest',
      {
        headers: {
          Accept: 'application/vnd.github.v3+json'
        }
      }
    )

    const j = await r.json()
    const latestVersion = (j.tag_name || '').replace(/^v/, '')
    const currentVersion = appVersion.value

    if (compareVersions(currentVersion, latestVersion) < 0) {
      appUpdateMsg.value = `Update available: ${latestVersion}`
      BrowserOpenURL(
        j.html_url || 'https://github.com/MiladGolchinpour/swiftray/releases/latest'
      )
    } else {
      appUpdateMsg.value = `You're up to date (${currentVersion}).`
    }
  } catch {
    appUpdateMsg.value = 'Failed to check for updates.'
  }

  checkingAppUpdate.value = false
}

function openGitHub() { BrowserOpenURL('https://github.com/MiladGolchinpour/swiftray') }

// --- Theme ---
const currentTheme = ref<'light' | 'dark' | 'system'>('dark')
const themeOptions = [{ label: 'Light', value: 'light' as const, icon: Sun }, { label: 'Dark', value: 'dark' as const, icon: Moon }, { label: 'System', value: 'system' as const, icon: Monitor }]
function setTheme(mode: 'light' | 'dark' | 'system') { currentTheme.value = mode; localStorage.setItem('swiftray-theme', mode); const root = document.documentElement; if (mode === 'light') root.classList.remove('dark'); else if (mode === 'dark') root.classList.add('dark'); else root.classList.toggle('dark', window.matchMedia('(prefers-color-scheme: dark)').matches) }
onMounted(() => { const saved = localStorage.getItem('swiftray-theme') as 'light' | 'dark' | 'system' | null; if (saved) setTheme(saved) })
</script>
