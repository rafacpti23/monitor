<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Servidores</h1>
        <p class="page-subtitle">{{ servers.length }} servidor(es) cadastrado(s)</p>
      </div>
      <button class="btn btn-primary" @click="$router.push('/add-server')">
        <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 4v12M4 10h12"/>
        </svg>
        Adicionar Servidor
      </button>
    </div>

    <div class="card" style="padding: 0; overflow: hidden;">
      <div v-if="loading" class="empty-state">Carregando...</div>
      <table v-else-if="servers.length" class="data-table">
        <thead>
          <tr>
            <th>Servidor</th>
            <th>OS</th>
            <th>CPU</th>
            <th>RAM</th>
            <th>Disco</th>
            <th>Load (1m)</th>
            <th>Uptime</th>
            <th>Último Seen</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="server in servers" :key="server.id" :class="{ 'row-stale': isStale(server) }" @click="$router.push(`/servers/${server.id}`)" style="cursor: pointer;">
            <td>
              <div style="display: flex; align-items: center; gap: 10px;">
                <div class="os-icon">{{ server.os === 'windows' ? '🪟' : '🐧' }}</div>
                <div>
                  <div style="font-weight: 500;">{{ server.name }}</div>
                  <div style="font-size: 11px; color: var(--text-muted); font-family: var(--font-mono);">{{ server.hostname || '—' }}</div>
                </div>
              </div>
            </td>
            <td><span class="badge badge-info">{{ server.os || '—' }}</span></td>
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
            <td><span class="mono" style="font-size: 12px;">{{ server.latest_metrics?.load1 ?? '—' }}</span></td>
            <td><span class="mono" style="font-size: 12px; color: var(--text-secondary);">{{ uptime(server) }}</span></td>
            <td><span class="mono" style="font-size: 11px; color: var(--text-muted);">{{ timeAgo(server.last_seen) }}</span></td>
            <td>
              <span v-if="isStale(server)" class="badge badge-danger" :title="'Última leitura ' + timeAgo(server.last_seen)">
                ○ offline
              </span>
              <span v-else class="badge badge-success">
                ● online
              </span>
              <div v-if="isStale(server)" class="stale-note">sem dados há {{ timeAgo(server.last_seen) }}</div>
            </td>
            <td>
              <button class="btn btn-ghost btn-sm" @click.stop="removeServer(server)">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M4 4h12M8 4v12M12 4v12M3 4l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">Nenhum servidor ainda. Clique em "Adicionar Servidor" para começar.</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import api from '../composables/useApi'
import { useSocket } from '../composables/useSocket'

const loading = ref(true)
const servers = ref([])
const { on: wsOn } = useSocket()

async function loadServers(showLoading = true) {
  if (showLoading) loading.value = true
  try {
    const res = await api.get('/servers')
    servers.value = Array.isArray(res.data) ? res.data : []
  } catch {
    servers.value = []
  } finally {
    if (showLoading) loading.value = false
  }
}

function applyUpdate(p) {
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
}

function applyOffline(p) {
  if (!p || !p.server_id) return
  const s = servers.value.find(x => x.id === p.server_id)
  if (s) s.status = 'offline'
}

let pollTimer = null

async function removeServer(server) {
  if (!confirm(`Remover servidor "${server.name}"?`)) return
  try {
    await api.delete(`/servers/${server.id}`)
    await loadServers()
  } catch { /* noop */ }
}

function pct(used, total) {
  if (!total) return 0
  return Math.round((used / total) * 100)
}
function cpu(s) { return Math.round(s.latest_metrics?.cpu_percent ?? 0) }
function ram(s) { return pct(s.latest_metrics?.ram_used, s.latest_metrics?.ram_total) }
function disk(s) { return pct(s.latest_metrics?.disk_used, s.latest_metrics?.disk_total) }

function uptime(s) {
  const sec = s.latest_metrics?.uptime_seconds
  if (!sec) return '—'
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  return `${d}d ${h}h`
}

function isStale(s) {
  if (s.status !== 'online') return true
  if (!s.last_seen) return true
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
  if (!dateStr) return 'nunca'
  const t = new Date(dateStr).getTime()
  if (!t) return '—'
  const diff = Math.floor((Date.now() - t) / 1000)
  if (diff < 60) return 'agora'
  if (diff < 3600) return `${Math.floor(diff / 60)}m atrás`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h atrás`
  return `${Math.floor(diff / 86400)}d atrás`
}

onMounted(async () => {
  await loadServers()
  wsOn('server_update', applyUpdate)
  wsOn('server_offline', applyOffline)
  pollTimer = setInterval(() => loadServers(false), 30000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.mono { font-family: var(--font-mono); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }
.row-stale { opacity: 0.55; filter: grayscale(0.7); }
.row-stale .bar-fill { background: var(--text-muted) !important; }
.row-stale .bar-value { color: var(--text-muted) !important; text-decoration: line-through; }
.stale-note { font-size: 10px; color: var(--accent-red); font-family: var(--font-mono); margin-top: 2px; }
</style>
