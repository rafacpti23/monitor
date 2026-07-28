import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../composables/useApi'

export const useIncidentsStore = defineStore('incidents', () => {
  const incidents = ref([])
  const activeCount = ref(0)
  const loading = ref(false)

  async function fetchAll() {
    loading.value = true
    try {
      const res = await api.get('/incidents')
      incidents.value = res.data || []
      activeCount.value = incidents.value.filter(i => !i.resolved).length
    } finally {
      loading.value = false
    }
  }

  async function acknowledge(id) {
    const res = await api.put(`/incidents/${id}/acknowledge`)
    const inc = incidents.value.find(i => i.id === id)
    if (inc) {
      inc.acknowledged = res.data.acknowledged
    }
  }

  async function resolve(id) {
    const res = await api.put(`/incidents/${id}/resolve`)
    const inc = incidents.value.find(i => i.id === id)
    if (inc) {
      inc.resolved = true
      activeCount.value = Math.max(0, activeCount.value - 1)
    }
  }

  return { incidents, activeCount, loading, fetchAll, acknowledge, resolve }
})
