import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'

export const useWebsitesStore = defineStore('websites', () => {
  const api = useApi()
  
  const websites = ref([])
  const currentWebsite = ref(null)
  const loading = ref(false)
  const error = ref(null)
  
  const websitesOnline = computed(() => {
    return websites.value.filter(w => w.status === 'up').length
  })
  
  const totalWebsites = computed(() => websites.value.length)
  
  const avgUptime = computed(() => {
    if (websites.value.length === 0) return 0
    const sum = websites.value.reduce((acc, w) => acc + (w.uptime_percent || 0), 0)
    return (sum / websites.value.length).toFixed(1)
  })
  
  async function fetchWebsites() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/api/v1/websites')
      websites.value = response.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }
  
  async function fetchWebsite(id) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/api/v1/websites/${id}`)
      currentWebsite.value = response.data
      return response.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }
  
  async function createWebsite(data) {
    const response = await api.post('/api/v1/websites', data)
    websites.value.push(response.data)
    return response.data
  }
  
  async function updateWebsite(id, data) {
    const response = await api.put(`/api/v1/websites/${id}`, data)
    const index = websites.value.findIndex(w => w.id === id)
    if (index !== -1) {
      websites.value[index] = response.data
    }
    if (currentWebsite.value?.id === id) {
      currentWebsite.value = response.data
    }
    return response.data
  }
  
  async function deleteWebsite(id) {
    await api.delete(`/api/v1/websites/${id}`)
    websites.value = websites.value.filter(w => w.id !== id)
    if (currentWebsite.value?.id === id) {
      currentWebsite.value = null
    }
  }
  
  async function getWebsiteHistory(id, params = {}) {
    const response = await api.get(`/api/v1/websites/${id}/history`, { params })
    return response.data
  }
  
  function updateWebsiteStatus(id, status, responseTime) {
    const website = websites.value.find(w => w.id === id)
    if (website) {
      website.status = status
      if (responseTime !== undefined) {
        website.response_time_ms = responseTime
      }
      website.last_checked = new Date().toISOString()
    }
    if (currentWebsite.value?.id === id) {
      currentWebsite.value.status = status
      if (responseTime !== undefined) {
        currentWebsite.value.response_time_ms = responseTime
      }
      currentWebsite.value.last_checked = new Date().toISOString()
    }
  }
  
  return {
    websites,
    currentWebsite,
    loading,
    error,
    websitesOnline,
    totalWebsites,
    avgUptime,
    fetchWebsites,
    fetchWebsite,
    createWebsite,
    updateWebsite,
    deleteWebsite,
    getWebsiteHistory,
    updateWebsiteStatus
  }
})
