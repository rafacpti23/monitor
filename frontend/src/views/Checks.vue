<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Checks</h1>
        <p class="page-subtitle">{{ checks.length }} verificação(ões) · Ping, TCP, HTTP, DNS, SSL</p>
      </div>
      <button class="btn btn-primary" @click="openAdd">
        <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 4v12M4 10h12"/>
        </svg>
        Adicionar Check
      </button>
    </div>

    <div class="card" style="padding: 0; overflow: hidden;">
      <div v-if="loading" class="empty-state">Carregando...</div>
      <table v-else-if="checks.length" class="data-table">
        <thead>
          <tr>
            <th>Nome</th>
            <th>Tipo</th>
            <th>Alvo</th>
            <th>Status</th>
            <th>Última Verificação</th>
            <th>Resposta</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="check in checks" :key="check.id">
            <td style="font-weight: 500;">{{ check.name }}</td>
            <td><span class="badge" :class="typeBadge(check.type)">{{ (check.type || '').toUpperCase() }}</span></td>
            <td class="font-mono text-muted" style="font-size: 11px;">{{ displayTarget(check) }}</td>
            <td><StatusBadge :status="check.status" /></td>
            <td class="font-mono text-muted" style="font-size: 11px;">{{ timeAgo(check.last_checked) }}</td>
            <td class="font-mono">{{ check.last_response_time_ms ?? 0 }}ms</td>
            <td>
              <button class="btn btn-ghost btn-sm" @click="deleteCheck(check)">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4h12M8 4v12M12 4v12"/></svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">Nenhum check cadastrado ainda.</div>
    </div>

    <Modal :visible="showAddModal" title="Adicionar Check" @close="showAddModal = false">
      <form @submit.prevent="createCheck">
        <div class="form-group">
          <label class="form-label">Nome do Check</label>
          <input v-model="form.name" class="form-input" placeholder="ex: Ping do Google DNS" required />
        </div>
        <div class="form-group">
          <label class="form-label">Tipo</label>
          <select v-model="form.type" class="form-select">
            <option value="ping">Ping (ICMP)</option>
            <option value="tcp">TCP Port</option>
            <option value="http">HTTP</option>
            <option value="dns">DNS</option>
            <option value="ssl_expiry">SSL Expiry</option>
          </select>
        </div>

        <div v-if="form.type === 'ping'" class="form-group">
          <label class="form-label">Host</label>
          <input v-model="form.target" class="form-input" placeholder="8.8.8.8" required />
        </div>

        <template v-if="form.type === 'tcp'">
          <div class="form-group">
            <label class="form-label">Host</label>
            <input v-model="form.target" class="form-input" placeholder="db.exemplo.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">Porta</label>
            <input v-model.number="form.port" type="number" class="form-input" placeholder="5432" required />
          </div>
        </template>

        <template v-if="form.type === 'http'">
          <div class="form-group">
            <label class="form-label">URL</label>
            <input v-model="form.url" class="form-input" placeholder="https://api.exemplo.com/health" required />
          </div>
          <div class="form-group">
            <label class="form-label">Código Esperado</label>
            <input v-model.number="form.expected_code" type="number" class="form-input" placeholder="200" />
          </div>
        </template>

        <template v-if="form.type === 'dns'">
          <div class="form-group">
            <label class="form-label">Hostname</label>
            <input v-model="form.hostname" class="form-input" placeholder="exemplo.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">IP Esperado (opcional)</label>
            <input v-model="form.expected_ip" class="form-input" placeholder="1.2.3.4" />
          </div>
        </template>

        <template v-if="form.type === 'ssl_expiry'">
          <div class="form-group">
            <label class="form-label">Hostname</label>
            <input v-model="form.hostname" class="form-input" placeholder="exemplo.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">Avisar (dias antes de expirar)</label>
            <input v-model.number="form.warn_days" type="number" class="form-input" placeholder="30" />
          </div>
        </template>

        <div v-if="error" class="alert alert-danger" style="margin-bottom: 12px;">{{ error }}</div>
        <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
          <button type="button" class="btn btn-secondary" @click="showAddModal = false">Cancelar</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">
            {{ saving ? 'Salvando...' : 'Adicionar' }}
          </button>
        </div>
      </form>
    </Modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Modal from '../components/Modal.vue'
import StatusBadge from '../components/StatusBadge.vue'
import api from '../composables/useApi'

const loading = ref(true)
const showAddModal = ref(false)
const saving = ref(false)
const error = ref('')
const checks = ref([])

function emptyForm() {
  return { name: '', type: 'ping', target: '', port: 80, url: '', expected_code: 200, hostname: '', expected_ip: '', warn_days: 30 }
}
const form = ref(emptyForm())

async function loadChecks() {
  loading.value = true
  try {
    const res = await api.get('/checks')
    checks.value = Array.isArray(res.data) ? res.data : []
  } catch {
    checks.value = []
  } finally {
    loading.value = false
  }
}

function typeBadge(type) {
  const map = { ping: 'badge-info', tcp: 'badge-info', http: 'badge-success', dns: 'badge-warning', ssl_expiry: 'badge-neutral' }
  return map[type] || 'badge-neutral'
}

function displayTarget(c) {
  if (c.type === 'ping') return c.target
  if (c.type === 'tcp') return `${c.target}:${c.port}`
  if (c.type === 'http') return c.url
  if (c.type === 'dns') return c.hostname
  if (c.type === 'ssl_expiry') return `${c.hostname} (${c.warn_days || 30}d)`
  return c.target || c.url || c.hostname || ''
}

function openAdd() {
  form.value = emptyForm()
  error.value = ''
  showAddModal.value = true
}

async function createCheck() {
  saving.value = true
  error.value = ''
  try {
    await api.post('/checks', form.value)
    showAddModal.value = false
    await loadChecks()
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  } finally {
    saving.value = false
  }
}

async function deleteCheck(check) {
  if (!confirm(`Deletar "${check.name}"?`)) return
  try {
    await api.delete(`/checks/${check.id}`)
    await loadChecks()
  } catch { /* noop */ }
}

function timeAgo(dateStr) {
  if (!dateStr) return 'nunca'
  const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diff < 60) return 'agora'
  if (diff < 3600) return `${Math.floor(diff / 60)}m atrás`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h atrás`
  return `${Math.floor(diff / 86400)}d atrás`
}

onMounted(loadChecks)
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.text-muted { color: var(--text-muted); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }
</style>
