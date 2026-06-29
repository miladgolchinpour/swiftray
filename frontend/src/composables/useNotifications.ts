import { ref } from 'vue'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

const notifications = ref<Notification[]>([])

let nextId = 0

function add(type: Notification['type'], title: string, message?: string, duration = 5000) {
  const id = String(++nextId)
  const notification: Notification = { id, type, title, message, duration }
  notifications.value.push(notification)

  if (duration > 0) {
    setTimeout(() => {
      remove(id)
    }, duration)
  }

  return id
}

function remove(id: string) {
  const idx = notifications.value.findIndex(n => n.id === id)
  if (idx !== -1) {
    notifications.value.splice(idx, 1)
  }
}

function clear() {
  notifications.value = []
}

function success(title: string, message?: string) {
  return add('success', title, message)
}

function error(title: string, message?: string) {
  return add('error', title, message, 8000)
}

function warning(title: string, message?: string) {
  return add('warning', title, message, 6000)
}

function info(title: string, message?: string) {
  return add('info', title, message)
}

function handleAPIError(errorMsg: string, context: string) {
  error(context, errorMsg)
}

export function useNotifications() {
  return {
    notifications,
    add,
    remove,
    clear,
    success,
    error,
    warning,
    info,
    handleAPIError,
  }
}
