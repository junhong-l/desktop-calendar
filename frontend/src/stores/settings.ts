import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '@/wailsjs/go/app/App'

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<any | null>(null)
  const loading = ref(false)

  async function fetchSettings() {
    loading.value = true
    try {
      settings.value = await api.GetSettings()
    } catch (error) {
      console.error('Failed to fetch settings:', error)
    } finally {
      loading.value = false
    }
  }

  async function updateSettings(newSettings: any) {
    try {
      await api.UpdateSettings(newSettings as any)
      settings.value = newSettings
    } catch (error) {
      console.error('Failed to update settings:', error)
      throw error
    }
  }

  return {
    settings,
    loading,
    fetchSettings,
    updateSettings
  }
})
