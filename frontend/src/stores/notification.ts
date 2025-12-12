import { defineStore } from 'pinia'
import { ref } from 'vue'

interface NotificationData {
  todo: any
  currentCount: number
  totalCount: number
  message: string
}

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<NotificationData[]>([])
  const unreadCount = ref(0)

  function addNotification(data: NotificationData) {
    notifications.value.unshift(data)
    unreadCount.value++
  }

  function clearNotifications() {
    notifications.value = []
    unreadCount.value = 0
  }

  function markAsRead() {
    unreadCount.value = 0
  }

  return {
    notifications,
    unreadCount,
    addNotification,
    clearNotifications,
    markAsRead
  }
})
