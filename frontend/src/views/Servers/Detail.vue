<template>
  <div v-if="loading" class="empty-state">Carregando servidor...</div>
  <div v-else-if="!server" class="empty-state">Servidor não encontrado.</div>
  <div v-else>
    <!-- Header -->
    <div class="page-header" style="margin-bottom: 24px;">
      <div style="display: flex; gap: 16px; align-items: center;">
        <button class="btn btn-ghost" @click="$router.push('/servers')" style="padding: 4px;">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 4L6 10l8 6"/>
          </svg>
        </button>
        <div>
          <div style="display: flex; align-items: center; gap: 12px;">
            <h1 class="page-title">{{ server.name }}</h1>
            <StatusBadge :status="server.status" />
            <span class="badge badge-info">{{ server.os || '—' }}</span>
          </div>
          <p class="page-subtitle font-mono">{{ server.hostname || '—' }} · Uptime: {{ uptimeStr }} · Último seen: {{ timeAgo(server.last_seen) }}</p>
        </div>
      </div>
    </div>

    <!-- Top Metrics -->
    <div class="metrics-grid">
      <div class="card" style="padding: 0;">
        <DonutMetric :value="cpuPct" label="CPU" color="var(--accent)" />
      </div>
      <div class="card" style="padding: 0;">
        <DonutMetric :value="ramPct" label="Memória" color="var(--accent-blue)" />
      </div>
      <div class="card" style="padding: 0;">
        <DonutMetric :value="diskPct" label="Disco" color="var(--accent-purple)" />
      </div>
    </div>

    <div class="stats-grid" style="margin-bottom: 24px;">
      <div class="metric-card">
        <div class="metric-label">Load (1/5/15)</div>
        <div class="metric-value font-mono" style="font-size: 20px; margin-top: 8px;">{{ loadStr }}</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Network I/O</div>
        <div class="metric-value font-mono" style="margin-top: 8px; display: flex; flex-direction: column; gap: 4px; font-size: 14px;">
          <div style="color: var(--accent-blue);">↑ {{ fmtBytes(m?.net_tx) }}</div>
          <div style="color: var(--accent);">↓ {{ fmtBytes(m?.net_rx) }}</div>
        </div>
      </div>
      <div class="metric-card">
        <div class="metric-label">RAM Total</div>
        <div class="metric-value font-mono" style="font-size: 20px; margin-top: 8px;">{{ fmtBytes(m?.ram_total) }}</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">Serviços</div>
        <div class="metric-value font-mono" style="font-size: 20px; margin-top: 8px;">{{ (server.services || []).length }}</div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="card" style="padding: 0;">
      <div class="tabs-header" style="display: flex; border-bottom: 1px solid var(--border); overflow-x: auto;">
        <button v-for="tab in tabs" :key="tab.id" class="tab-btn" :class="{ active: currentTab === tab.id }" @click="currentTab = tab.id">
          {{ tab.label }}
        </button>
      </div>

      <div class="tab-content" style="padding: 20px;">
        <div v-if="currentTab === 'overview'">
          <h3 style="font-size: 14px; font-weight: 600; margin-bottom: 16px;">Uso de CPU</h3>
          <Chart v-if="cpuHistory.length" :data="cpuHistory" color="var(--accent)" :height="250" />
          <div v-else class="empty-state">Sem histórico ainda. Aguardando dados do agente.</div>
        </div>

        <div v-if="currentTab === 'services'">
          <table v-if="(server.services || []).length" class="data-table">
            <thead><tr><th>Serviço</th><th>Status</th></tr></thead>
            <tbody>
              <tr v-for="s in server.services" :key="s.id || s.name">
                <td style="font-weight: 500;">{{ s.name }}</td>
                <td><StatusBadge :status="s.status" /></td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state">Nenhum serviço reportado pelo agente.</div>
        </div>

        <div v-if="currentTab === 'incidents'">
          <table v-if="incidents.length" class="data-table">
            <thead><tr><th>Tipo</th><th>Mensagem</th><th>Severidade</th><th>Criado</th></tr></thead>
            <tbody>
              <tr v-for="i in incidents" :key="i.id">
                <td><span class="badge badge-info">{{ i.type }}</span></td>
                <td>{{ i.message }}</td>
                <td>{{ i.severity }}</td>
                <td class="font-mono" style="font-size: 11px;">{{ timeAgo(i.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state">Sem incidentes para este servidor.</div>
        </div>

        <div v-if="currentTab === 'settings'">
          <div style="max-width: 500px;">
            <div class="form-group">
              <label class="form-label">Nome do Servidor</label>
              <input v-model="editName" class="form-input" />
            </div>
            <button class="btn btn-primary" :disabled="saving" @click="saveServer">
              {{ saving ? 'Salvando...' : 'Salvar' }}
            </button>
            <div style="margin-top: 32px; padding-top: 24px; border-top: 1px solid var(--border);">
              <h4 style="color: var(--accent-red); margin-bottom: 16px;">Zona de Perigo</h4>
              <button class="btn btn-danger" @click="confirmDelete = true">Deletar Servidor</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ConfirmDialog
      :visible="confirmDelete"
      title="Deletar Servidor?"
      message="Tem certeza? Todos os históricos e métricas deste servidor serão permanentemente removidos."
      confirmText="Deletar"
      @close="confirmDelete = false"
      @confirm="deleteServer"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Chart from '../../components/Chart.vue'
import DonutMetric from '../../components/DonutMetric.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import ConfirmDialog from '../../components/ConfirmDialog.vue'
import api from '../../composables/useApi'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const server = ref(null)
const incidents = ref([])
const cpuHistory = ref([])
const editName = ref('')
const saving = ref(false)
const confirmDelete = ref(false)
const currentTab = ref('overview')

const tabs = [
  { id: 'overview', label: 'Overview' },
  { id: 'services', label: 'Serviços' },
  { id: 'incidents', label: 'Incidentes' },
  { id: 'settings', label: 'Settings' }
]

const m = computed(() => server.value?.latest_metrics)
const cpuPct = computed(() => Math.round(m.value?.cpu_percent ?? 0))
const ramPct = computed(() => pct(m.value?.ram_used, m.value?.ram_total))
const diskPct = computed(() => pct(m.value?.disk_used, m.value?.disk_total))
const loadStr = computed(() => {
  if (!m.value) return '—'
  return `${m.value.load1 ?? 0}, ${m.value.load5 ?? 0}, ${m.value.load15 ?? 0}`
})
const uptimeStr = computed(() => {
  const sec = m.value?.uptime_seconds
  if (!sec) return '—'
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  return `${d}d ${h}h`
})

function pct(used, total) {
  if (!total) return 0
  return Math.round((used / total) * 100)
}

function fmtBytes(b) {
  if (!b) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0, val = b
  while (val >= 1024 && i < units.length - 1) { val /= 1024; i++ }
  return `${val.toFixed(1)} ${units[i]}`
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

async function loadServer() {
  loading.value = true
  try {
    const res = await api.get(`/servers/${route.params.id}`)
    server.value = res.data
    editName.value = res.data?.name || ''
    // Fetch history + incidents in parallel (best-effort)
    const [histRes, incRes] = await Promise.all([
      api.get(`/servers/${route.params.id}/history`).catch(() => ({ data: [] })),
      api.get(`/servers/${route.params.id}/incidents`).catch(() => ({ data: [] })),
    ])
    const hist = Array.isArray(histRes.data) ? histRes.data : []
    cpuHistory.value = hist.map((h, i) => ({ x: i, y: h.cpu_percent ?? 0 }))
    incidents.value = Array.isArray(incRes.data) ? incRes.data : []
  } catch {
    server.value = null
  } finally {
    loading.value = false
  }
}

async function saveServer() {
  saving.value = true
  try {
    await api.put(`/servers/${route.params.id}`, { name: editName.value })
    await loadServer()
  } catch { /* noop */ } finally {
    saving.value = false
  }
}

async function deleteServer() {
  confirmDelete.value = false
  try {
    await api.delete(`/servers/${route.params.id}`)
  } catch { /* noop */ }
  router.push('/servers')
}

onMounted(loadServer)
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }
.tab-btn {
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-family: var(--font-body);
  font-size: 13.5px;
  font-weight: 500;
  padding: 12px 20px;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition-fast);
}
.tab-btn:hover { color: var(--text-primary); }
.tab-btn.active { color: var(--accent); border-bottom-color: var(--accent); }
</style>
