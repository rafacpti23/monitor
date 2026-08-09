<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Gestor WhatsApp</h1>
        <p class="page-subtitle">
          {{ panels.length }} painel(is) monitorado(s)
          <span v-if="totalInstances > 0" class="text-muted">
            · {{ connectedInstances }}/{{ totalInstances }} instâncias conectadas
          </span>
        </p>
      </div>
      <button class="btn btn-primary" @click="openAdd">
        <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 4v12M4 10h12"/>
        </svg>
        Adicionar Painel
      </button>
    </div>

    <div v-if="loading" class="card"><div class="empty-state">Carregando...</div></div>

    <div v-else-if="!panels.length" class="card">
      <div class="empty-state">
        Nenhum painel monitorado ainda.<br/>
        Adicione o primeiro para acompanhar suas instâncias de WhatsApp!
      </div>
    </div>

    <div v-else class="panels">
      <div v-for="panel in panels" :key="panel.id" class="card panel-card">
        <div class="panel-head" @click="toggle(panel.id)">
          <div class="panel-info">
            <div class="panel-status-dot" :class="panelDotClass(panel)"></div>
            <div>
              <div class="panel-name">
                {{ panel.name }}
                <span class="provider-badge" :class="'provider-' + (panel.provider || 'papi')">
                  {{ (panel.provider || 'papi').toUpperCase() }}
                </span>
              </div>
              <div class="panel-url font-mono text-muted">{{ panel.base_url }}</div>
            </div>
          </div>
          <div class="panel-meta">
            <div class="panel-count">
              <span :class="allConnected(panel) ? 'ok' : 'warn'">
                {{ panel.connected_instances }}/{{ panel.total_instances }}
              </span>
              <span class="text-muted" style="font-size: 11px;">conectadas</span>
            </div>
            <div class="panel-checked font-mono text-muted">{{ timeAgo(panel.last_checked) }}</div>
            <div class="panel-actions">
              <button class="btn btn-ghost btn-sm" @click.stop="checkNow(panel)" title="Verificar agora">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M16 4v4h-4M4 16v-4h4"/>
                  <path d="M16 8a6 6 0 0 0-11-2M4 12a6 6 0 0 0 11 2"/>
                </svg>
              </button>
              <button class="btn btn-ghost btn-sm" @click.stop="openEdit(panel)" title="Editar">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2l4 4L7 17l-5 1 1-5z"/>
                </svg>
              </button>
              <button class="btn btn-ghost btn-sm" @click.stop="deletePanel(panel)" title="Excluir">
                <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M4 4h12M8 4v12M12 4v12M3 4l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12"/>
                </svg>
              </button>
              <svg class="chevron" :class="{ open: expanded[panel.id] }" width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M6 8l4 4 4-4"/>
              </svg>
            </div>
          </div>
        </div>

        <div v-if="panel.status === 'error'" class="panel-error font-mono">
          ⚠ Erro ao conectar: {{ panel.last_error }}
        </div>

        <div v-if="expanded[panel.id]" class="panel-body">
          <div v-if="loadingInstances[panel.id]" class="empty-state">Carregando instâncias...</div>
          <table v-else-if="(instances[panel.id] || []).length" class="data-table">
            <thead>
              <tr>
                <th>Instância</th>
                <th>Nome</th>
                <th>Número</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="inst in instances[panel.id]" :key="inst.id" :class="{ 'row-down': !isConnected(inst) }">
                <td class="font-mono" style="font-size: 11px;">{{ inst.instance_id }}</td>
                <td>{{ inst.name || '—' }}</td>
                <td class="font-mono text-muted">{{ inst.phone_number || '—' }}</td>
                <td>
                  <span class="inst-badge" :class="isConnected(inst) ? 'connected' : 'disconnected'">
                    {{ isConnected(inst) ? '● CONNECTED' : '○ ' + (inst.status || 'UNKNOWN') }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="empty-state">Nenhuma instância retornada por este painel.</div>
        </div>
      </div>
    </div>

    <Modal :visible="showModal" :title="editingId ? 'Editar Painel' : 'Adicionar Painel'" @close="showModal = false">
      <form @submit.prevent="savePanel">
        <div class="form-group">
          <label class="form-label">Provedor</label>
          <div class="provider-select">
            <button type="button" class="provider-option" :class="{ active: form.provider === 'papi' }"
                    @click="selectProvider('papi')">
              <span class="provider-icon">🔗</span>
              <span class="provider-name">PAPI</span>
              <span class="provider-desc">papi.api.br</span>
            </button>
            <button type="button" class="provider-option" :class="{ active: form.provider === 'stevo' }"
                    @click="selectProvider('stevo')">
              <span class="provider-icon">💬</span>
              <span class="provider-name">Stevo</span>
              <span class="provider-desc">stevo.chat</span>
            </button>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">Nome</label>
          <input v-model="form.name" class="form-input" :placeholder="form.provider === 'stevo' ? 'ex: Stevo Produção' : 'ex: PAPI Produção'" required />
        </div>
        <div class="form-group">
          <label class="form-label">{{ form.provider === 'stevo' ? 'URL base da API' : 'URL base do painel' }}</label>
          <input v-model="form.base_url" class="form-input" :placeholder="form.provider === 'stevo' ? 'https://openapi.stevo.chat' : 'https://papi.api.br'" />
          <div class="form-hint">Sem barra no final.{{ form.provider === 'stevo' ? ' O P-mon chama <base>/mcp' : ' O P-mon chama <base>/api/v1/instances' }}</div>
        </div>
        <div class="form-group">
          <label class="form-label">
            {{ form.provider === 'stevo' ? 'Token de acesso (Bearer)' : 'Token global do painel (x-panel-token)' }}
            <span v-if="editingId" class="text-muted" style="font-weight: 400;">— deixe em branco p/ manter</span>
          </label>
          <input v-model="form.panel_token" type="password" class="form-input"
                 :placeholder="editingId ? '•••••••• (não alterar)' : 'SEU_TOKEN'"
                 :required="!editingId" />
        </div>
        <div class="form-group">
          <label class="form-label">Intervalo de checagem</label>
          <select v-model.number="form.check_interval_sec" class="form-select">
            <option :value="30">30 segundos</option>
            <option :value="60">1 minuto</option>
            <option :value="120">2 minutos</option>
            <option :value="300">5 minutos</option>
          </select>
          <div class="form-hint">De quanto em quanto tempo o painel é verificado. Alerta dispara quando uma instância não está CONNECTED.</div>
        </div>
        <div v-if="error" class="alert alert-danger" style="margin-bottom: 12px;">{{ error }}</div>
        <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
          <button type="button" class="btn btn-secondary" @click="showModal = false">Cancelar</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">
            {{ saving ? 'Salvando...' : (editingId ? 'Salvar' : 'Adicionar') }}
          </button>
        </div>
      </form>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import Modal from '../components/Modal.vue'
import api from '../composables/useApi'

const loading = ref(true)
const panels = ref([])
const expanded = ref({})
const instances = ref({})
const loadingInstances = ref({})
const showModal = ref(false)
const saving = ref(false)
const error = ref('')
const editingId = ref(null)

let pollTimer = null

function emptyForm() {
  return { name: '', provider: 'papi', base_url: 'https://papi.api.br', panel_token: '', check_interval_sec: 60 }
}
const form = ref(emptyForm())

function selectProvider(provider) {
  form.value.provider = provider
  if (provider === 'stevo') {
    form.value.base_url = 'https://openapi.stevo.chat'
  } else {
    form.value.base_url = 'https://papi.api.br'
  }
}

const totalInstances = computed(() => panels.value.reduce((s, p) => s + (p.total_instances || 0), 0))
const connectedInstances = computed(() => panels.value.reduce((s, p) => s + (p.connected_instances || 0), 0))

function isConnected(inst) {
  const s = (inst.status || '').toUpperCase()
  return s === 'CONNECTED' || s === 'OPEN' || s === 'ACTIVE'
}
function allConnected(panel) {
  return panel.total_instances > 0 && panel.connected_instances === panel.total_instances
}
function panelDotClass(panel) {
  if (panel.status === 'error') return 'error'
  if (panel.status === 'pending') return 'pending'
  return allConnected(panel) ? 'ok' : 'warn'
}

async function loadPanels(showLoading = true) {
  if (showLoading) loading.value = true
  try {
    const res = await api.get('/papi/panels')
    panels.value = Array.isArray(res.data) ? res.data : []
    // Refresh open panels' instances.
    for (const id of Object.keys(expanded.value)) {
      if (expanded.value[id]) loadInstances(Number(id), false)
    }
  } catch {
    panels.value = []
  } finally {
    loading.value = false
  }
}

async function loadInstances(panelId, showLoading = true) {
  if (showLoading) loadingInstances.value = { ...loadingInstances.value, [panelId]: true }
  try {
    const res = await api.get(`/papi/panels/${panelId}`)
    instances.value = { ...instances.value, [panelId]: res.data.instances || [] }
  } catch {
    instances.value = { ...instances.value, [panelId]: [] }
  } finally {
    loadingInstances.value = { ...loadingInstances.value, [panelId]: false }
  }
}

function toggle(panelId) {
  const open = !expanded.value[panelId]
  expanded.value = { ...expanded.value, [panelId]: open }
  if (open && !instances.value[panelId]) loadInstances(panelId)
}

function openAdd() {
  form.value = emptyForm()
  editingId.value = null
  error.value = ''
  showModal.value = true
}

function openEdit(panel) {
  form.value = {
    name: panel.name,
    provider: panel.provider || 'papi',
    base_url: panel.base_url,
    panel_token: '',
    check_interval_sec: panel.check_interval_sec || 60,
  }
  editingId.value = panel.id
  error.value = ''
  showModal.value = true
}

async function savePanel() {
  saving.value = true
  error.value = ''
  try {
    if (editingId.value) {
      await api.put(`/papi/panels/${editingId.value}`, form.value)
    } else {
      await api.post('/papi/panels', form.value)
    }
    showModal.value = false
    setTimeout(() => loadPanels(false), 1500) // give the async first-check a moment
    await loadPanels(false)
  } catch (e) {
    error.value = e.response?.data?.error || e.message
  } finally {
    saving.value = false
  }
}

async function checkNow(panel) {
  try {
    await api.post(`/papi/panels/${panel.id}/check`)
    setTimeout(() => { loadPanels(false); if (expanded.value[panel.id]) loadInstances(panel.id, false) }, 1500)
  } catch { /* noop */ }
}

async function deletePanel(panel) {
  if (!confirm(`Deletar o painel "${panel.name}"? Isso remove o monitoramento das instâncias dele.`)) return
  try {
    await api.delete(`/papi/panels/${panel.id}`)
    await loadPanels(false)
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

onMounted(() => {
  loadPanels()
  pollTimer = setInterval(() => loadPanels(false), 30000)
})
onUnmounted(() => { if (pollTimer) clearInterval(pollTimer) })
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.text-muted { color: var(--text-muted); }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; line-height: 1.6; }

.panels { display: flex; flex-direction: column; gap: 12px; }
.panel-card { padding: 0; overflow: hidden; }

.panel-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px; cursor: pointer; transition: background 0.15s;
}
.panel-head:hover { background: rgba(255,255,255,0.02); }

.panel-info { display: flex; align-items: center; gap: 14px; }
.panel-status-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.panel-status-dot.ok { background: var(--accent, #00e676); box-shadow: 0 0 8px rgba(0,230,118,0.5); }
.panel-status-dot.warn { background: #ffb300; box-shadow: 0 0 8px rgba(255,179,0,0.5); }
.panel-status-dot.error { background: var(--accent-red, #f44336); box-shadow: 0 0 8px rgba(244,67,54,0.5); }
.panel-status-dot.pending { background: var(--text-muted, #8b949e); }

.panel-name { font-weight: 600; font-size: 14px; }
.panel-url { font-size: 11px; margin-top: 2px; }

.panel-meta { display: flex; align-items: center; gap: 20px; }
.panel-count { display: flex; flex-direction: column; align-items: flex-end; }
.panel-count .ok { color: var(--accent, #00e676); font-weight: 600; font-family: var(--font-mono); }
.panel-count .warn { color: #ffb300; font-weight: 600; font-family: var(--font-mono); }
.panel-checked { font-size: 11px; min-width: 70px; text-align: right; }
.panel-actions { display: flex; align-items: center; gap: 4px; }
.chevron { transition: transform 0.2s; color: var(--text-muted); }
.chevron.open { transform: rotate(180deg); }

.panel-error {
  padding: 8px 20px; background: rgba(244,67,54,0.08);
  color: var(--accent-red, #f44336); font-size: 12px;
  border-top: 1px solid var(--border, #21262d);
}

.panel-body { border-top: 1px solid var(--border, #21262d); }
.row-down { background: rgba(244,67,54,0.05); }

.inst-badge { font-family: var(--font-mono); font-size: 11px; font-weight: 600; }
.inst-badge.connected { color: var(--accent, #00e676); }
.inst-badge.disconnected { color: var(--accent-red, #f44336); }

/* Provider badge on panel cards */
.provider-badge {
  display: inline-block; font-size: 10px; font-weight: 700; padding: 1px 6px;
  border-radius: 4px; margin-left: 8px; vertical-align: middle; letter-spacing: 0.5px;
}
.provider-badge.provider-papi { background: rgba(0,230,118,0.15); color: var(--accent, #00e676); }
.provider-badge.provider-stevo { background: rgba(33,150,243,0.15); color: #2196f3; }

/* Provider selector in modal */
.provider-select { display: flex; gap: 10px; }
.provider-option {
  flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 14px 12px; border: 1.5px solid var(--border, #21262d); border-radius: 8px;
  background: transparent; color: var(--text); cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); font-family: inherit;
}
.provider-option:hover { border-color: rgba(255,255,255,0.15); background: rgba(255,255,255,0.03); }
.provider-option.active { border-color: var(--accent, #00e676); background: rgba(0,230,118,0.06); }
.provider-icon { font-size: 20px; }
.provider-name { font-weight: 600; font-size: 13px; }
.provider-desc { font-size: 11px; color: var(--text-muted); font-family: var(--font-mono); }
</style>
