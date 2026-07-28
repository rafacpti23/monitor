import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../composables/useApi'

export const useServersStore = defineStore('servers', () => {
  const servers = ref([])
  const currentServer = ref(null)
  const loading = ref(false)

  async function fetchAll() {
    loading.value = true
    try {
      const res = await api.get('/servers')
      servers.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  async function fetchOne(id) {
    loading.value = true
    try {
      const res = await api.get(`/servers/${id}`)
      currentServer.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function createServer(payload) {
    const res = await api.post('/servers', payload)
    servers.value.push(res.data)
    return res.data
  }

  function updateMetric(serverId, metrics) {
    const srv = servers.value.find(s => s.id === serverId)
    if (srv) {
      Object.assign(srv, metrics)
    }
    if (currentServer.value && currentServer.value.id === serverId) {
      Object.assign(currentServer.value, metrics)
    }
  }

  return { servers, currentServer, loading, fetchAll, fetchOne, createServer, updateMetric }
})
