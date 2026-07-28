<template>
  <div>
    <div v-if="loading" style="color: var(--text-muted); padding: 40px; text-align: center;">Carregando...</div>

    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-else-if="site">
      <div class="page-header" style="margin-bottom: 24px;">
        <div style="display: flex; gap: 16px; align-items: center;">
          <button class="btn btn-ghost" @click="$router.push('/websites')" style="padding: 4px;">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M14 4L6 10l8 6"/>
            </svg>
          </button>
          <div>
            <div style="display: flex; align-items: center; gap: 12px;">
              <h1 class="page-title">{{ site.name }}</h1>
              <StatusBadge :status="site.status || 'pending'" />
            </div>
            <p class="page-subtitle font-mono">{{ site.url }} · Última checagem: {{ site.last_checked ? timeAgo(site.last_checked) : 'nunca' }}</p>
          </div>
        </div>
        <div style="display: flex; gap: 8px;">
          <button class="btn btn-secondary btn-sm" @click="openEdit">Editar</button>
          <button class="btn btn-secondary btn-sm" @click="loadAll(true)">Atualizar</button>
          <button class="btn btn-danger btn-sm" @click="deleteSite">Excluir</button>
        </div>
      </div>

      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">Status atual</div>
          <div class="metric-value" style="font-size: 24px; margin-top: 12px;"
               :style="{ color: statusColor }">
            {{ statusLabel }}
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Último response</div>
          <div class="metric-value font-mono" style="font-size: 24px; margin-top: 12px;">
            {{ site.last_response_code || '—' }}
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Tempo de resposta</div>
          <div class="metric-value font-mono" style="font-size: 24px; margin-top: 12px;">
            {{ site.last_response_time_ms ? site.last_response_time_ms + 'ms' : '—' }}
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Uptime (últimas {{ history.length || 0 }} checagens)</div>
          <div class="metric-value font-mono" style="font-size: 24px; margin-top: 12px;"
               :style="{ color: uptime >= 99 ? 'var(--accent)' : uptime >= 95 ? 'var(--accent-yellow, #ffc107)' : 'var(--accent-red, #f44336)' }">
            {{ uptime !== null ? uptime.toFixed(2) + '%' : '—' }}
          </div>
        </div>
      </div>

      <!-- Chart -->
      <div class="card" style="margin-bottom: 24px;">
        <div class="card-header"><h3 class="card-title">Tempo de Resposta ({{ history.length }} checagens)</h3></div>
        <div v-if="history.length === 0" style="padding: 40px; text-align: center; color: var(--text-muted);">
          Sem histórico ainda. Aguarde a próxima checagem.
        </div>
        <Chart v-else :data="chartData" color="var(--accent-blue, #42a5f5)" :height="250" />
      </div>

      <!-- Configuração -->
      <div class="card" style="margin-bottom: 24px;">
        <div class="card-header"><h3 class="card-title">Configuração</h3></div>
        <div class="config-grid">
          <div><span class="cfg-label">Método:</span> <span class="font-mono">{{ site.method || 'GET' }}</span></div>
          <div><span class="cfg-label">Intervalo:</span> <span class="font-mono">{{ site.check_interval_sec || 60 }}s</span></div>
          <div v-if="site.search_string"><span class="cfg-label">Search string:</span> <span class="font-mono">{{ site.search_string }}</span></div>
          <div v-if="site.ssl_expires_at"><span class="cfg-label">SSL expira:</span> <span class="font-mono">{{ formatDate(site.ssl_expires_at) }}</span></div>
          <div><span class="cfg-label">Criado em:</span> <span class="font-mono">{{ formatDate(site.created_at) }}</span></div>
        </div>
      </div>

      <!-- Incidentes recentes -->
      <div class="card" style="margin-bottom: 24px;">
        <div class="card-header"><h3 class="card-title">Incidentes Recentes</h3></div>
        <div v-if="incidents.length === 0" style="padding: 20px; color: var(--text-muted); font-size: 13px;">
          Nenhum incidente registrado 🎉
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Horário</th>
              <th>Tipo</th>
              <th>Severidade</th>
              <th>Status</th>
              <th>Resolvido em</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inc in incidents" :key="inc.id">
              <td class="font-mono">{{ formatDate(inc.start_time) }}</td>
              <td>{{ inc.message || inc.alert_type }}</td>
              <td>
                <span class="badge" :class="inc.severity === 'critical' ? 'badge-danger' : 'badge-warning'">
                  {{ inc.severity }}
                </span>
              </td>
              <td>
                <span class="badge" :class="inc.status === 'resolved' ? 'badge-success' : 'badge-danger'">
                  {{ inc.status === 'resolved' ? 'Resolvido' : 'Ativo' }}
                </span>
              </td>
              <td class="font-mono">{{ inc.resolved_at ? formatDate(inc.resolved_at) : '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Últimas checagens (15 mais recentes) -->
      <div class="card">
        <div class="card-header"><h3 class="card-title">Últimas 15 checagens</h3></div>
        <div v-if="history.length === 0" style="padding: 20px; color: var(--text-muted); font-size: 13px;">
          Nenhuma checagem registrada.
        </div>
        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Horário</th>
              <th>Código HTTP</th>
              <th>Tempo</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in history.slice(0, 15)" :key="h.id">
              <td class="font-mono">{{ formatDate(h.timestamp) }}</td>
              <td class="font-mono">{{ h.response_code }}</td>
              <td class="font-mono">{{ h.response_time_ms }}ms</td>
              <td>
                <span :style="{ color: h.status_ok ? 'var(--accent)' : 'var(--accent-red, #f44336)' }">
                  {{ h.status_ok ? '✓ OK' : '✗ FALHA' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Edit Modal -->
      <Modal v-if="showEditModal" :visible="showEditModal" title="Editar Website" @close="showEditModal = false">
        <form @submit.prevent="saveEdit">
          <div class="form-group">
            <label class="form-label">Nome</label>
            <input v-model="editForm.name" class="form-input" required />
          </div>
          <div class="form-group">
            <label class="form-label">URL</label>
            <input v-model="editForm.url" type="url" class="form-input" required />
          </div>
          <div class="form-group">
            <label class="form-label">Intervalo de checagem</label>
            <select v-model.number="editForm.check_interval_sec" class="form-select">
              <option :value="30">30 segundos</option>
              <option :value="60">1 minuto</option>
              <option :value="120">2 minutos</option>
              <option :value="300">5 minutos</option>
              <option :value="600">10 minutos</option>
              <option :value="1800">30 minutos</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">Procurar String</label>
            <input v-model="editForm.search_string" class="form-input" placeholder="ex: Bem-vindo" />
          </div>
          <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
            <button type="button" class="btn btn-secondary" @click="showEditModal = false">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="editSaving">
              {{ editSaving ? 'Salvando...' : 'Salvar' }}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../composables/useApi'
import Chart from '../../components/Chart.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import Modal from '../../components/Modal.vue'

const route = useRoute()
const router = useRouter()

const site = ref(null)
const history = ref([])
const incidents = ref([])
const loading = ref(true)
const error = ref('')
const showEditModal = ref(false)
const editSaving = ref(false)
const editForm = ref({})
let pollTimer = null

const statusLabel = computed(() => {
  if (!site.value) return '—'
  const s = site.value.status
  if (s === 'online' || s === 'up') return 'Online'
  if (s === 'offline' || s === 'down') return 'Offline'
  if (s === 'pending') return 'Aguardando'
  return s || '—'
})

const statusColor = computed(() => {
  if (!site.value) return 'var(--text-muted)'
  const s = site.value.status
  if (s === 'online' || s === 'up') return 'var(--accent)'
  if (s === 'offline' || s === 'down') return 'var(--accent-red, #f44336)'
  return 'var(--text-muted)'
})

const uptime = computed(() => {
  if (history.value.length === 0) return null
  const ok = history.value.filter(h => h.status_ok).length
  return (ok / history.value.length) * 100
})

const chartData = computed(() =>
  // Chart component provavelmente aceita array de {x, y}; ordena por timestamp asc
  [...history.value]
    .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
    .map((h, i) => ({ x: i, y: h.response_time_ms || 0 }))
)

async function loadAll(showLoading = false) {
  if (showLoading || !site.value) loading.value = true
  error.value = ''
  try {
    const [siteRes, histRes, incRes] = await Promise.all([
      api.get(`/websites/${route.params.id}`),
      api.get(`/websites/${route.params.id}/history`).catch(() => ({ data: [] })),
      api.get(`/websites/${route.params.id}/incidents`).catch(() => ({ data: [] })),
    ])
    site.value = siteRes.data
    history.value = Array.isArray(histRes.data) ? histRes.data : []
    incidents.value = Array.isArray(incRes.data) ? incRes.data : []
  } catch (e) {
    error.value = e.response?.data?.error || 'Falha ao carregar website.'
  } finally {
    loading.value = false
  }
}

function openEdit() {
  if (!site.value) return
  editForm.value = {
    name: site.value.name,
    url: site.value.url,
    check_interval_sec: site.value.check_interval_sec || 60,
    search_string: site.value.search_string || '',
  }
  showEditModal.value = true
}

async function saveEdit() {
  editSaving.value = true
  try {
    await api.put(`/websites/${route.params.id}`, editForm.value)
    showEditModal.value = false
    await loadAll(true)
  } catch (e) {
    alert(e.response?.data?.error || 'Falha ao salvar.')
  } finally {
    editSaving.value = false
  }
}

async function deleteSite() {
  if (!confirm(`Excluir website "${site.value.name}"?`)) return
  try {
    await api.delete(`/websites/${route.params.id}`)
    router.push('/websites')
  } catch (e) {
    alert(e.response?.data?.error || 'Falha ao excluir.')
  }
}

function formatDate(d) {
  if (!d) return '—'
  const date = new Date(d)
  return date.toLocaleString('pt-BR')
}

function timeAgo(d) {
  if (!d) return 'nunca'
  const diffMs = Date.now() - new Date(d).getTime()
  const s = Math.floor(diffMs / 1000)
  if (s < 60) return `${s}s atrás`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}min atrás`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h atrás`
  return `${Math.floor(h / 24)}d atrás`
}

onMounted(async () => {
  await loadAll(true)
  // Auto-refresh a cada 30s (sem spinner).
  pollTimer = setInterval(() => loadAll(false), 30000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px 24px;
  padding: 4px 0;
  font-size: 13px;
}
.cfg-label { color: var(--text-muted); margin-right: 6px; }
</style>
