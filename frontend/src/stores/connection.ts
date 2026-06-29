import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { api } from '@/api/client'
import { useNotifications } from '@/composables/useNotifications'

export type ConnectionState = 'idle' | 'connecting' | 'connected' | 'disconnecting' | 'error'

export const useConnectionStore = defineStore('connection', () => {
  const notify = useNotifications()

  const state = ref<ConnectionState>('idle')
  const errorMessage = ref<string>('')
  const statusMessage = ref<string>('')

  const isConnected = computed(() => state.value === 'connected')
  const isConnecting = computed(() => state.value === 'connecting')
  const isDisconnecting = computed(() => state.value === 'disconnecting')
  const isError = computed(() => state.value === 'error')
  const isBusy = computed(() => state.value === 'connecting' || state.value === 'disconnecting')
  const canConnect = computed(() => state.value === 'idle' || state.value === 'error')
  const canDisconnect = computed(() => state.value === 'connected')

  function startListening() {
    EventsOn('connection:state-change', (dataStr: string) => {
      try {
        const data = JSON.parse(dataStr)
        console.log('[Connection]', data.state, data.message || '', data.error || '')
        state.value = data.state
        statusMessage.value = data.message || ''
        if (data.error) {
          errorMessage.value = data.error
          notify.error('Connection failed', data.error)
        } else if (data.state === 'connected') {
          errorMessage.value = ''
          // Don't show "Connected" notification during settings reload
          // The app store handles that notification
        } else if (data.state === 'idle' && statusMessage.value === 'Disconnected') {
          errorMessage.value = ''
        }
      } catch (e) {
        console.error('[Connection] Failed to parse state event:', e)
      }
    })
  }

  async function syncState() {
    const { data } = await api.getConnectionState()
    if (data) {
      state.value = data.state as ConnectionState
      if (data.error) {
        errorMessage.value = data.error
      }
    }
  }

  async function connect() {
    if (!canConnect.value) return
    errorMessage.value = ''
    const { error } = await api.connect()
    if (error) {
      state.value = 'error'
      errorMessage.value = error
    }
  }

  async function disconnect() {
    if (!canDisconnect.value) return
    const { error } = await api.disconnect()
    if (error) {
      state.value = 'error'
      errorMessage.value = error
    }
  }

  return {
    state,
    errorMessage,
    statusMessage,
    isConnected,
    isConnecting,
    isDisconnecting,
    isError,
    isBusy,
    canConnect,
    canDisconnect,
    startListening,
    syncState,
    connect,
    disconnect,
  }
})
