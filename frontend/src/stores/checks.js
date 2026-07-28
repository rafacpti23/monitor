import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'

export const useChecksStore = defineStore('checks', () => {
  const api = useApi()
  
  const checks = ref([])
  const loading = ref(false)
  const error = ref(null)
  
  const checksByStatus = computed(() => {
    return {
      passing: checks.value.filter(c => c.status === 'passing').length,
      failing: checks.value.filter(c => c.status === 'failing').length,
      unknown: checks.value.filter(c => c.status === 'unknown').length
    }
  })
  
  async function fetchChecks() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/api/v1/checks')
      checks.value = response.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }
  
  async function createCheck(data) {
    const response = await api.post('/api/v1/checks', data)
    checks.value.push(response.data)
    return response.data
  }
  
  async function updateCheck(id, data) {
    const response = await api.put(`/api/v1/checks/${id}`, data)
    const index = checks.value.findIndex(c => c.id === id)
    if (index !== -1) {
      checks.value[index] = response.data
    }
    return response.data
  }
  
  async function deleteCheck(id) {
    await api.delete(`/api/v1/checks/${id}`)
    checks.value = checks.value.filter(c => c.id !== id)
  }
  
  function updateCheckStatus(id, status, result) {
    const check = checks.value.find(c => c.id === id)
    if (check) {
      check.status = status
      check.last_result = result
      check.last_checked = new Date().toISOString()
    }
  }
  
  return {
    checks,
    loading,
    error,
    checksByStatus,
    fetchChecks,
    createCheck,
    updateCheck,
    deleteCheck,
    updateCheckStatus
  }
})
