<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Websites</h1>
        <p class="page-subtitle">{{ websites.length }} monitoramento(s) HTTP/HTTPS</p>
      </div>
      <button class="btn btn-primary" @click="openAdd">
        <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 4v12M4 10h12"/>
        </svg>
        Adicionar Website
      </button>
    </div>

    <div class="card" style="padding: 0; overflow: hidden;">
      <div v-if="loading" class="empty-state">Carregando...</div>
      <table v-else-if="websites.length" class="data-table">
        <thead>
          <tr>
            <th>Website</th>
            <th>URL</th>
            <th>Status</th>
            <th>HTTP</th>
            <th>Resp Time</th>
            <th>Última Checagem</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="site in websites" :key="site.id" @click="$router.push(`/websites/${site.id}`)">
            <td style="font-weight: 500;">{{ site.name }}</td>
            <td class="font-mono text-muted" style="font-size: 11px;">{{ site.url }}</td>
            <td><StatusBadge :status="site.status" /></td>
            <td class="font-mono">{{ site.last_response_code || '—' }}</td>
            <td class="font-mono">{{ site.last_response_time_ms ?? 0 }}ms</td>
            <td class="font-mono text-muted" style="font-size: 11px;">{{ timeAgo(site.last_checked) }}</td>
            <td style="white-space: nowrap;">
              <button class="btn btn-ghost btn-sm" @click.stop="openEdit(site)" title="Editar">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2l4 4L7 17l-5 1 1-5z"/>
                </svg>
              </button>
              <button class="btn btn-ghost btn-sm" @click.stop="deleteSite(site)" title="Excluir">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M4 4h12M8 4v12M12 4v12M3 4l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">Nenhum website monitorado ainda. Adicione o primeiro!</div>
    </div>

    <Modal :visible="showAddModal" :title="editingId ? 'Editar Website' : 'Adicionar Website'" @close="showAddModal = false">
      <form @submit.prevent="saveWebsite">
        <div class="form-group">
          <label class="form-label">Nome</label>
          <input v-model="form.name" class="form-input" placeholder="ex: Site Principal" required />
        </div>
        <div class="form-group">
          <label class="form-label">URL</label>
          <input v-model="form.url" type="url" class="form-input" placeholder="https://..." required />
        </div>
        <div class="form-group">
          <label class="form-label">Intervalo de checagem</label>
          <select v-model.number="form.check_interval_sec" class="form-select">
            <option :value="30">30 segundos</option>
            <option :value="60">1 minuto</option>
            <option :value="120">2 minutos</option>
            <option :value="300">5 minutos</option>
            <option :value="600">10 minutos</option>
            <option :value="1800">30 minutos</option>
          </select>
          <div class="form-hint">De quanto em quanto tempo o site é verificado.</div>
        </div>
        <div class="form-group">
          <label class="form-label">Procurar String (opcional)</label>
          <input v-model="form.search_string" class="form-input" placeholder="ex: Bem-vindo" />
          <div class="form-hint">O site será considerado offline se não contiver este texto.</div>
        </div>
        <div v-if="error" class="alert alert-danger" style="margin-bottom: 12px;">{{ error }}</div>
        <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
          <button type="button" class="btn btn-secondary" @click="showAddModal = false">Cancelar</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">
            {{ saving ? 'Salvando...' : (editingId ? 'Salvar' : 'Adicionar') }}
          </button>
        </div>
      </form>
    </Modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import StatusBadge from '../components/StatusBadge.vue'
import Modal from '../components/Modal.vue'
import api from '../composables/useApi'

const loading = ref(true)
const websites = ref([])
const showAddModal = ref(false)
const saving = ref(false)
const error = ref('')
const editingId = ref(null)

function emptyForm() {
  return { name: '', url: '', check_interval_sec: 60, search_string: '' }
}
const form = ref(emptyForm())

async function loadWebsites() {
  loading.value = true
  try {
    const res = await api.get('/websites')
    websites.value = Array.isArray(res.data) ? res.data : []
  } catch {
    websites.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  form.value = emptyForm()
  editingId.value = null
  error.value = ''
  showAddModal.value = true
}

function openEdit(site) {
  form.value = {
    name: site.name,
    url: site.url,
    check_interval_sec: site.check_interval_sec || 60,
    search_string: site.search_string || '',
  }
  editingId.value = site.id
  error.value = ''
  showAddModal.value = true
}

async function saveWebsite() {
  saving.value = true
  error.value = ''
  try {
    if (editingId.value) {
      await api.put(`/websites/${editingId.value}`, form.value)
    } else {
      await api.post('/websites', form.value)
    }
    showAddModal.value = false
    await loadWebsites()
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  } finally {
    saving.value = false
  }
}

async function deleteSite(site) {
  if (!confirm(`Deletar "${site.name}"?`)) return
  try {
    await api.delete(`/websites/${site.id}`)
    await loadWebsites()
  } catch { /* noop */ }
}

function timeAgo(dateStr) {
  if (!dateStr) return 'nunca'
  const t = new Date(dateStr).getTime()
  if (!t) return '—'
  const diff = Math.floor((Date.now() - t) / 1000)
  if (diff < 60) return 'agora'
  if (diff < 3600) return `${Math.floor(diff / 60)}m atrás`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h atrás`
  return `${Math.floor(diff / 86400)}d atrás`
}

onMounted(loadWebsites)
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.text-muted { color: var(--text-muted); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }
</style>
