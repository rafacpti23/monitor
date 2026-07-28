<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Incidentes</h1>
        <p class="page-subtitle">{{ filtered.length }} incidente(s) · {{ activeCount }} ativo(s)</p>
      </div>
    </div>

    <!-- Filter Chips -->
    <div class="filter-chips">
      <button v-for="f in filters" :key="f.id" class="chip" :class="{ active: currentFilter === f.id }" @click="currentFilter = f.id">
        <span v-if="f.dot" class="dot" :class="f.dot"></span>
        {{ f.label }}
        <span class="chip-count">{{ countByFilter(f.id) }}</span>
      </button>
    </div>

    <div v-if="loading" class="card empty-state">Carregando...</div>

    <!-- Timeline grouped by day -->
    <template v-else>
      <div v-for="group in groupedIncidents" :key="group.day" class="day-group">
        <div class="day-label">{{ group.day }}</div>
        <div class="card" style="padding: 0;">
          <div v-for="inc in group.items" :key="inc.id" class="incident-row">
            <div class="incident-header" @click="toggleExpand(inc.id)">
              <div class="incident-dot" :class="inc.severity"></div>
              <div style="flex: 1;">
                <div style="display: flex; align-items: center; gap: 10px;">
                  <span style="font-weight: 500;">{{ inc.monitor_name || '—' }}</span>
                  <span class="badge" :class="`badge-${badgeType(inc.type)}`">{{ inc.type || '—' }}</span>
                  <span v-if="inc.resolved_at" class="badge badge-success">Resolvido</span>
                  <span v-else-if="inc.acknowledged_at" class="badge badge-warning">ACK</span>
                  <span v-else class="badge badge-danger">Ativo</span>
                </div>
                <div style="color: var(--text-secondary); font-size: 12.5px; margin-top: 4px;">{{ inc.message }}</div>
                <div class="incident-meta font-mono" style="margin-top: 4px;">
                  {{ formatTime(inc.created_at) }} → {{ inc.resolved_at ? formatTime(inc.resolved_at) : 'em andamento' }}
                </div>
              </div>
              <div class="incident-actions" @click.stop>
                <button v-if="!inc.acknowledged_at && !inc.resolved_at" class="btn btn-secondary btn-sm" @click="ack(inc)">ACK</button>
                <button v-if="!inc.resolved_at" class="btn btn-primary btn-sm" @click="resolve(inc)">Resolver</button>
                <button v-if="!inc.resolved_at" class="btn btn-ghost btn-sm" @click="ignore(inc)">Ignorar</button>
                <button class="btn btn-ghost btn-sm" @click="toggleExpand(inc.id)">
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" :style="{ transform: expanded.includes(inc.id) ? 'rotate(180deg)' : 'none' }">
                    <path d="M6 8l4 4 4-4"/>
                  </svg>
                </button>
              </div>
            </div>
            <div v-if="expanded.includes(inc.id)" class="incident-details">
              <h4 style="font-size: 12px; font-weight: 600; margin-bottom: 8px;">Detalhes</h4>
              <div class="metric-snapshot font-mono">
                <div><span class="text-muted">Severidade:</span> {{ inc.severity }}</div>
                <div><span class="text-muted">Tipo:</span> {{ inc.type || '—' }}</div>
                <div><span class="text-muted">Criado:</span> {{ inc.created_at }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div v-if="!loading && filtered.length === 0" class="card" style="text-align: center; color: var(--text-muted); padding: 60px;">
      Sem incidentes neste filtro. 🎉
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../composables/useApi'

const filters = [
  { id: 'all', label: 'Todos' },
  { id: 'active', label: 'Ativos', dot: 'critical' },
  { id: 'acknowledged', label: 'Reconhecidos', dot: 'warning' },
  { id: 'resolved', label: 'Resolvidos', dot: 'resolved' },
  { id: 'critical', label: 'Críticos', dot: 'critical' },
  { id: 'warning', label: 'Avisos', dot: 'warning' }
]

const currentFilter = ref('all')
const expanded = ref([])
const loading = ref(true)
const incidents = ref([])

async function loadIncidents() {
  loading.value = true
  try {
    const res = await api.get('/incidents')
    incidents.value = Array.isArray(res.data) ? res.data : []
  } catch {
    incidents.value = []
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  const f = currentFilter.value
  return incidents.value.filter(i => {
    if (f === 'all') return true
    if (f === 'active') return !i.resolved_at && !i.acknowledged_at
    if (f === 'acknowledged') return i.acknowledged_at && !i.resolved_at
    if (f === 'resolved') return !!i.resolved_at
    if (f === 'critical') return i.severity === 'critical'
    if (f === 'warning') return i.severity === 'warning'
    return true
  })
})

const activeCount = computed(() => incidents.value.filter(i => !i.resolved_at).length)

const groupedIncidents = computed(() => {
  const groups = {}
  filtered.value.forEach(inc => {
    const day = dayLabel(inc.created_at)
    if (!groups[day]) groups[day] = []
    groups[day].push(inc)
  })
  return Object.entries(groups).map(([day, items]) => ({ day, items }))
})

function countByFilter(id) {
  return incidents.value.filter(i => {
    if (id === 'all') return true
    if (id === 'active') return !i.resolved_at && !i.acknowledged_at
    if (id === 'acknowledged') return i.acknowledged_at && !i.resolved_at
    if (id === 'resolved') return !!i.resolved_at
    if (id === 'critical') return i.severity === 'critical'
    if (id === 'warning') return i.severity === 'warning'
    return true
  }).length
}

function badgeType(t) {
  const map = { cpu: 'info', ram: 'info', docker: 'warning', nodata: 'danger', http: 'danger', service: 'warning' }
  return map[(t || '').toLowerCase()] || 'neutral'
}

function dayLabel(dateStr) {
  if (!dateStr) return 'Desconhecido'
  const d = new Date(dateStr)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) return 'Hoje'
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return 'Ontem'
  return d.toLocaleDateString('pt-BR')
}

function formatTime(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
}

function toggleExpand(id) {
  const i = expanded.value.indexOf(id)
  if (i >= 0) expanded.value.splice(i, 1)
  else expanded.value.push(id)
}

async function ack(inc) {
  try {
    await api.put(`/incidents/${inc.id}/acknowledge`)
    await loadIncidents()
  } catch { /* noop */ }
}

async function resolve(inc) {
  try {
    await api.put(`/incidents/${inc.id}/resolve`)
    await loadIncidents()
  } catch { /* noop */ }
}

async function ignore(inc) {
  try {
    await api.put(`/incidents/${inc.id}/ignore`)
    await loadIncidents()
  } catch { /* noop */ }
}

onMounted(loadIncidents)
</script>

<style scoped>
.filter-chips {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.chip:hover { border-color: var(--border-bright); color: var(--text-primary); }
.chip.active { background: var(--accent-glow); border-color: var(--accent); color: var(--accent); }
.chip .dot { width: 8px; height: 8px; border-radius: 50%; }
.chip .dot.critical { background: var(--accent-red); }
.chip .dot.warning { background: var(--accent-amber); }
.chip .dot.resolved { background: var(--accent); }
.chip-count {
  background: var(--bg-base);
  padding: 1px 6px;
  border-radius: 8px;
  font-family: var(--font-mono);
  font-size: 10px;
}

.day-group { margin-bottom: 24px; }
.day-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  padding: 8px 4px;
  letter-spacing: 1px;
}

.incident-row + .incident-row { border-top: 1px solid var(--border); }
.incident-header {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  align-items: flex-start;
  cursor: pointer;
  transition: background var(--transition-fast);
}
.incident-header:hover { background: var(--bg-card-hover); }
.incident-actions { display: flex; gap: 4px; align-items: center; }
.incident-details {
  padding: 0 20px 16px 48px;
  border-top: 1px solid var(--border);
  padding-top: 16px;
  background: var(--bg-base);
}
.metric-snapshot { display: flex; gap: 24px; padding: 12px; background: var(--bg-card); border-radius: var(--radius-sm); margin-bottom: 16px; font-size: 12px;}
.text-muted { color: var(--text-muted); }
.font-mono { font-family: var(--font-mono); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }
</style>
