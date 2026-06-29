import { ref, readonly } from 'vue'
import { api } from '@/api/client'
import { useNotifications } from '@/composables/useNotifications'

const isReady = ref(false)
const isLoading = ref(true)
const xrayInstalled = ref(false)
const xrayVersion = ref('')
const geoIPReady = ref(false)
const geoSiteReady = ref(false)
const dataDir = ref('')
const resourceError = ref('')

export function useStartup() {
  const notify = useNotifications()

  async function initialize() {
    isLoading.value = true

    const { data: info, error: infoErr } = await api.getAppInfo()
    if (infoErr) {
      notify.error('Startup failed', infoErr)
      isLoading.value = false
      return
    }

    if (info) {
      xrayInstalled.value = info.xrayInstalled
      xrayVersion.value = info.xrayVersion
      geoIPReady.value = info.geoIPReady
      geoSiteReady.value = info.geoSiteReady
      dataDir.value = info.dataDir
    }

    const { error: readyErr } = await api.verifyReady()
    if (readyErr) {
      resourceError.value = readyErr
      notify.error('Runtime resources missing', readyErr)
      isLoading.value = false
      return
    }

    isReady.value = true
    isLoading.value = false
    console.log('[Startup] App initialized', info)
  }

  return {
    isReady: readonly(isReady),
    isLoading: readonly(isLoading),
    xrayInstalled: readonly(xrayInstalled),
    xrayVersion: readonly(xrayVersion),
    geoIPReady: readonly(geoIPReady),
    geoSiteReady: readonly(geoSiteReady),
    dataDir: readonly(dataDir),
    resourceError: readonly(resourceError),
    initialize,
  }
}
