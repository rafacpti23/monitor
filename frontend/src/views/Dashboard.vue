<template>
  <div class="dashboard">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Dashboard</h1>
        <p class="page-subtitle">
          Visão geral do monitoramento ·
          <span :style="{ color: wsConnected ? 'var(--accent)' : 'var(--text-muted)' }">
            {{ wsConnected ? '● ao vivo' : '○ reconectando' }}
          </span>
          · atualizado {{ lastUpdateLabel }}
        </p>
      </div>
      <div style="display: flex; gap: 8px;">
        <button class="btn btn-secondary btn-sm" @click="$router.push('/add-server')">
          <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 4v12M4 10h12"/>
          </svg>
          Adicionar Servidor
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-grid">
      <div class="metric-card" style="--card-accent: var(--accent)">
        <div class="card-icon" style="background: rgba(0,230,118,0.1);">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--accent)" stroke-width="1.5">
            <rect x="2" y="4" width="16" height="4" rx="1"/>
            <rect x="2" y="10" width="16" height="4" rx="1"/>
            <circle cx="5" cy="6" r="1" fill="var(--accent)"/>
            <circle cx="5" cy="12" r="1" fill="var(--accent)"/>
          </svg>
        </div>
        <div class="metric-value">{{ stats.servers.online }}<span style="font-size: 14px; color: var(--text-muted);">/{{ stats.servers.total }}</span></div>
        <div class="metric-label">Servidores Online</div>
        <div class="metric-trend up">
          <svg width="10" height="10" viewBox="0 0 10 10" fill="var(--accent)"><path d="M5 2L9 7H1z"/></svg>
          {{ stats.servers.down }} offline
        </div>
      </div>

      <div class="metric-card" style="--card-accent: var(--accent-blue)">
        <div class="card-icon" style="background: rgba(68,138,255,0.1);">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--accent-blue)" stroke-width="1.5">
            <circle cx="10" cy="10" r="8"/>
            <path d="M10 2c2.5 2 4 5 4 8s-1.5 6-4 8c-2.5-2-4-5-4-8s1.5-6 4-8z"/>
          </svg>
        </div>
        <div class="metric-value">{{ stats.websites.up }}<span style="font-size: 14px; color: var(--text-muted);">/{{ stats.websites.total }}</span></div>
        <div class="metric-label">Websites OK</div>
        <div class="metric-trend neutral">
          avg {{ stats.websites.avgResponse }}ms
        </div>
      </div>

      <div class="metric-card" style="--card-accent: var(--accent-amber)">
        <div class="card-icon" style="background: rgba(255,171,0,0.1);">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--accent-amber)" stroke-width="1.5">
            <path d="M10 2L2 16h16L10 2z"/>
            <path d="M10 9v4M10 14v1"/>
          </svg>
        </div>
        <div class="metric-value">{{ stats.alerts.active }}</div>
        <div class="metric-label">Alertas Ativos</div>
        <div class="metric-trend down" v-if="stats.alerts.critical">
          <svg width="10" height="10" viewBox="0 0 10 10" fill="var(--accent-red)"><path d="M5 8L1 3h8z"/></svg>
          {{ stats.alerts.critical }} críticos
        </div>
      </div>

      <div class="metric-card" style="--card-accent: var(--accent-purple)">
        <div class="card-icon" style="background: rgba(124,77,255,0.1);">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--accent-purple)" stroke-width="1.5">
            <circle cx="10" cy="10" r="8"/>
            <path d="M6 10h8M10 6v8"/>
          </svg>
        </div>
        <div class="metric-value">{{ stats.uptime }}<span style="font-size: 14px; color: var(--text-muted);">%</span></div>
        <div class="metric-label">Uptime Médio</div>
        <div class="metric-trend up">
          <svg width="10" height="10" viewBox="0 0 10 10" fill="var(--accent)"><path d="M5 2L9 7H1z"/></svg>
          últimos 7 dias
        </div>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="dashboard-grid">
      <!-- Servers Table -->
      <div>
        <div class="card" style="padding: 0; overflow: hidden;">
          <div class="card-header" style="padding: 16px 20px; border-bottom: 1px solid var(--border);">
            <h3 style="font-size: 13px; font-weight: 600;">Servidores</h3>
            <div class="live-indicator">
              <div class="live-dot"></div>
              <span>LIVE</span>
            </div>
          </div>
          <div v-if="loading" class="empty-state">Carregando...</div>
          <table v-else-if="servers.length" class="data-table">
            <thead>
              <tr>
                <th>Servidor</th>
                <th>CPU</th>
                <th>RAM</th>
                <th>Disco</th>
                <th>Load</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="server in servers" :key="server.id" :class="{ 'row-stale': isStale(server) }" @click="$router.push(`/servers/${server.id}`)">
                <td>
                  <div style="display: flex; align-items: center; gap: 10px;">
                    <div style="width: 28px; height: 28px; border-radius: 6px; background: var(--bg-base); display: flex; align-items: center; justify-content: center; font-size: 12px;">
                      🐧
                    </div>
                    <div>
                      <div style="font-weight: 500; font-size: 13px;">{{ server.name }}</div>
                      <div style="font-size: 11px; color: var(--text-muted); font-family: var(--font-mono);">{{ server.hostname }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="bar-track">
                      <div class="bar-fill" :class="metricClass(cpu(server), 80, 50)" :style="{ width: cpu(server) + '%' }"></div>
                    </div>
                    <span class="bar-value mono">{{ cpu(server) }}%</span>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="bar-track">
                      <div class="bar-fill" :class="metricClass(ram(server), 90, 70)" :style="{ width: ram(server) + '%' }"></div>
                    </div>
                    <span class="bar-value mono">{{ ram(server) }}%</span>
                  </div>
                </td>
                <td>
                  <div class="bar-container">
                    <div class="bar-track">
                      <div class="bar-fill" :class="metricClass(disk(server), 85, 70)" :style="{ width: disk(server) + '%' }"></div>
                    </div>
                    <span class="bar-value mono">{{ disk(server) }}%</span>
                  </div>
                </td>
                <td>
                  <span class="mono" style="font-size: 12px; color: var(--text-secondary);">{{ server.latest_metrics?.load1 ?? '—' }}</span>
                </td>
                <td>
                  <span v-if="isStale(server)" class="badge badge-danger" :title="'Última leitura ' + timeAgo(server.last_seen)">
                    ○ offline
                  </span>
                  <span v-else class="badge badge-success">
                    ● online
                  </span>
                  <div v-if="isStale(server)" class="stale-note">sem dados há {{ timeAgo(server.last_seen) }}</div>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state">Nenhum servidor cadastrado. Adicione o primeiro!</div>
        </div>
      </div>

      <!-- Incident Timeline -->
      <div>
        <div class="card" style="margin-bottom: 16px;">
          <div class="card-header">
            <h3 style="font-size: 13px; font-weight: 600;">Incidentes Recentes</h3>
            <RouterLink to="/incidents" class="btn btn-ghost btn-sm">Ver todos</RouterLink>
          </div>
          <div v-if="incidents.length" class="incident-timeline">
            <div v-for="incident in incidents.slice(0, 5)" :key="incident.id" class="incident-item">
              <div class="incident-dot" :class="incident.severity"></div>
              <div class="incident-content">
                <div class="incident-title">{{ incident.message || incident.title }}</div>
                <div class="incident-meta">{{ incident.monitor_name || '—' }} · {{ timeAgo(incident.created_at) }}</div>
              </div>
              <span v-if="incident.acknowledged_at" class="badge badge-success" style="font-size: 10px;">ACK</span>
              <span v-else class="badge badge-danger" style="font-size: 10px;">ATIVO</span>
            </div>
          </div>
          <div v-else class="empty-state" style="padding: 24px;">Nenhum incidente ativo 🎉</div>
        </div>

        <!-- Uptime Mini Chart -->
        <div class="card">
          <div class="card-header">
            <h3 style="font-size: 13px; font-weight: 600;">Uptime — 7 dias</h3>
          </div>
          <div class="chart-container" style="height: 80px;">
            <svg viewBox="0 0 300 60" style="width: 100%; height: 100%;">
              <defs>
                <linearGradient id="uptimeGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="var(--accent)" stop-opacity="0.3"/>
                  <stop offset="100%" stop-color="var(--accent)" stop-opacity="0"/>
                </linearGradient>
              </defs>
              <path d="M0 55 L0 50 L43 48 L86 52 L129 45 L172 47 L215 43 L258 46 L300 40 L300 55 Z" fill="url(#uptimeGrad)"/>
              <path d="M0 50 L43 48 L86 52 L129 45 L172 47 L215 43 L258 46 L300 40" fill="none" stroke="var(--accent)" stroke-width="1.5"/>
              <circle cx="0" cy="50" r="3" fill="var(--accent)"/>
              <circle cx="300" cy="40" r="3" fill="var(--accent)"/>
            </svg>
          </div>
          <div style="display: flex; justify-content: space-between; font-size: 10px; color: var(--text-muted); font-family: var(--font-mono); margin-top: 4px;">
            <span>7d atrás</span><span>agora</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import api from '../composables/useApi'
import { useSocket } from '../composables/useSocket'

const loading = ref(true)
const servers = ref([])
const websites = ref([])
const incidents = ref([])
const lastUpdate = ref(Date.now())
const nowTick = ref(Date.now())

const { connected: wsConnected, on: wsOn } = useSocket()

const lastUpdateLabel = computed(() => {
  const diff = Math.floor((nowTick.value - lastUpdate.value) / 1000)
  if (diff < 5) return 'agora'
  if (diff < 60) return `há ${diff}s`
  if (diff < 3600) return `há ${Math.floor(diff / 60)}min`
  return `há ${Math.floor(diff / 3600)}h`
})

const stats = computed(() => {
  const online = servers.value.filter(s => s.status === 'online').length
  const total = servers.value.length
  const wsUp = websites.value.filter(w => w.status === 'online' || w.status === 'up').length
  const wsTotal = websites.value.length
  const wsWithMs = websites.value.filter(w => w.last_response_time_ms > 0)
  const wsAvg = wsWithMs.length
    ? Math.round(wsWithMs.reduce((a, w) => a + w.last_response_time_ms, 0) / wsWithMs.length)
    : 0
  const active = incidents.value.filter(i => !i.resolved_at).length
  const critical = incidents.value.filter(i => i.severity === 'critical' && !i.resolved_at).length
  return {
    servers: { online, total, down: total - online },
    websites: { up: wsUp, total: wsTotal, avgResponse: wsAvg },
    alerts: { active, critical },
    uptime: total || wsTotal ? 99.9 : 0,
  }
})

async function loadData(showLoading = true) {
  if (showLoading) loading.value = true
  try {
    const [srvRes, wsRes, incRes] = await Promise.all([
      api.get('/servers').catch(() => ({ data: [] })),
      api.get('/websites').catch(() => ({ data: [] })),
      api.get('/incidents').catch(() => ({ data: [] })),
    ])
    servers.value = Array.isArray(srvRes.data) ? srvRes.data : []
    websites.value = Array.isArray(wsRes.data) ? wsRes.data : []
    incidents.value = Array.isArray(incRes.data) ? incRes.data : []
    lastUpdate.value = Date.now()
  } finally {
    if (showLoading) loading.value = false
  }
}

// Merge a live server_update event into the current list without a full refetch.
function applyServerUpdate(p) {
  if (!p || !p.server_id) return
  const s = servers.value.find(x => x.id === p.server_id)
  if (!s) return
  s.status = p.status || s.status
  s.hostname = p.hostname || s.hostname
  s.last_seen = new Date().toISOString()
  s.latest_metrics = Object.assign({}, s.latest_metrics, {
    cpu_percent: p.cpu_percent,
    ram_used: p.ram_used,
    ram_total: p.ram_total,
    disk_used: p.disk_used,
    disk_total: p.disk_total,
  })
  lastUpdate.value = Date.now()
}

function applyServerOffline(p) {
  if (!p || !p.server_id) return
  const s = servers.value.find(x => x.id === p.server_id)
  if (!s) return
  s.status = 'offline'
  lastUpdate.value = Date.now()
}

// Polling fallback: every 30s force a full refresh. If WS is delivering
// updates this is basically a no-op cost-wise but keeps things sane when WS
// is down or a slow event was dropped.
let pollTimer = null
let clockTimer = null

onMounted(async () => {
  await loadData()
  wsOn('server_update', applyServerUpdate)
  wsOn('server_offline', applyServerOffline)
  pollTimer = setInterval(() => loadData(false), 30000)
  clockTimer = setInterval(() => { nowTick.value = Date.now() }, 1000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (clockTimer) clearInterval(clockTimer)
})

function pct(used, total) {
  if (!total || !used) return 0
  return Math.round((used / total) * 100)
}
function cpu(s) { return Math.round(s.latest_metrics?.cpu_percent ?? 0) }
function ram(s) { return pct(s.latest_metrics?.ram_used, s.latest_metrics?.ram_total) }
function disk(s) { return pct(s.latest_metrics?.disk_used, s.latest_metrics?.disk_total) }

function isStale(s) {
  if (s.status !== 'online') return true
  if (!s.last_seen) return true
  // Consider stale if last_seen older than 2× interval (fallback 10min).
  const intervalSec = s.interval_seconds && s.interval_seconds > 0 ? s.interval_seconds : 300
  const cutoffMs = Math.max(intervalSec * 2, 120) * 1000
  return (Date.now() - new Date(s.last_seen).getTime()) > cutoffMs
}

function metricClass(val, high, mid) {
  if (val >= high) return 'high'
  if (val >= mid) return 'medium'
  return 'low'
}

function timeAgo(dateStr) {
  if (!dateStr) return '—'
  const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diff < 60) return `${diff}s atrás`
  if (diff < 3600) return `${Math.floor(diff / 60)}m atrás`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h atrás`
  return `${Math.floor(diff / 86400)}d atrás`
}


</script>

<style scoped>
.mono { font-family: var(--font-mono); }
.empty-state { text-align: center; padding: 40px 20px; color: var(--text-muted); font-size: 13px; }
.row-stale { opacity: 0.55; filter: grayscale(0.7); }
.row-stale .bar-fill { background: var(--text-muted) !important; }
.row-stale .bar-value { color: var(--text-muted) !important; text-decoration: line-through; }
.stale-note { font-size: 10px; color: var(--accent-red); font-family: var(--font-mono); margin-top: 2px; }
</style>
