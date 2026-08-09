<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">Configurações</h1>
        <p class="page-subtitle">Perfil, canais de alerta e sistema</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs-header">
      <button v-for="tab in tabs" :key="tab.id" class="tab-btn" :class="{ active: currentTab === tab.id }" @click="currentTab = tab.id">
        {{ tab.label }}
      </button>
    </div>

    <!-- PERFIL -->
    <div v-if="currentTab === 'profile'" class="card" style="max-width: 600px;">
      <div class="form-group">
        <label class="form-label">Nome</label>
        <input v-model="profile.name" class="form-input" />
      </div>
      <div class="form-group">
        <label class="form-label">Email</label>
        <input v-model="profile.email" type="email" class="form-input" disabled />
        <div class="form-hint">Email não pode ser alterado.</div>
      </div>
      <div class="form-group">
        <label class="form-label">Fuso Horário</label>
        <select v-model="profile.timezone" class="form-select">
          <option value="America/Sao_Paulo">America/Sao_Paulo (UTC-3)</option>
          <option value="America/New_York">America/New_York (UTC-5)</option>
          <option value="UTC">UTC</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label">WhatsApp</label>
        <input v-model="profile.whatsapp_number" class="form-input" placeholder="5511999999999" />
      </div>
      <div style="display: flex; align-items: center; gap: 12px;">
        <button class="btn btn-primary" :disabled="savingProfile" @click="saveProfile">
          {{ savingProfile ? 'Salvando...' : 'Salvar Perfil' }}
        </button>
        <span v-if="profileMsg" class="save-msg">{{ profileMsg }}</span>
      </div>
    </div>

    <!-- CANAIS DE ALERTA -->
    <div v-if="currentTab === 'channels'">
      <div v-if="loadingChannels" class="card empty-state">Carregando canais...</div>
      <template v-else>
        <!-- Each channel -->
        <div v-for="ch in channels" :key="ch.id" class="card" style="margin-bottom: 20px;">
          <div class="card-header">
            <h3 class="card-title">{{ ch.name || ch.type }}</h3>
            <label class="toggle">
              <input type="checkbox" v-model="ch.enabled" @change="updateChannel(ch)" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div style="color: var(--text-muted); font-size: 12px; margin-bottom: 12px;">
            Tipo: <strong>{{ ch.type }}</strong>
          </div>
          <div style="display: flex; gap: 8px;">
            <button class="btn btn-secondary btn-sm" @click="testChannel(ch)">Testar</button>
            <button class="btn btn-ghost btn-sm" @click="deleteChannel(ch)">Remover</button>
          </div>
        </div>

        <button class="btn btn-primary" @click="showAddChannel = true">+ Adicionar Canal</button>
      </template>

      <div v-if="testResult" class="alert" :class="testResult.ok ? 'alert-success' : 'alert-danger'" style="margin-top: 20px;">
        {{ testResult.message }}
      </div>

      <!-- Add Channel Modal -->
      <div v-if="showAddChannel" class="modal-overlay" @click.self="showAddChannel = false">
        <div class="modal">
          <div class="modal-header">
            <h3>Novo Canal de Alerta</h3>
            <button class="btn btn-ghost btn-sm" @click="showAddChannel = false">✕</button>
          </div>
          <form @submit.prevent="createChannel">
            <div class="form-group">
              <label class="form-label">Nome</label>
              <input v-model="channelForm.name" class="form-input" placeholder="Email Principal" required />
            </div>
            <div class="form-group">
              <label class="form-label">Tipo</label>
              <select v-model="channelForm.type" class="form-select">
                <option value="email">Email (SMTP)</option>
                <option value="whatsapp">WhatsApp (PAPI)</option>
                <option value="webhook">Webhook</option>
              </select>
            </div>

            <!-- Email config -->
            <template v-if="channelForm.type === 'email'">
              <div style="display: grid; grid-template-columns: 2fr 1fr; gap: 12px;">
                <div class="form-group">
                  <label class="form-label">Host SMTP</label>
                  <input v-model="channelForm.config.host" class="form-input" placeholder="smtp.gmail.com" />
                </div>
                <div class="form-group">
                  <label class="form-label">Porta</label>
                  <input v-model.number="channelForm.config.port" type="number" class="form-input" placeholder="587" />
                </div>
              </div>
              <div class="form-group">
                <label class="form-label">Usuário</label>
                <input v-model="channelForm.config.user" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Senha</label>
                <input v-model="channelForm.config.pass" type="password" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">De (From)</label>
                <input v-model="channelForm.config.from" class="form-input" placeholder="alertas@p-mon.com" />
              </div>
            </template>

            <!-- WhatsApp PAPI config -->
            <template v-if="channelForm.type === 'whatsapp'">
              <p class="provider-note">Provedor: <strong>PAPI</strong> — <code>api.papi.api.br</code></p>
              <div class="form-group">
                <label class="form-label">Instância</label>
                <input v-model="channelForm.config.instance" class="form-input" placeholder="minha_instancia" />
              </div>
              <div class="form-group">
                <label class="form-label">API Key</label>
                <input v-model="channelForm.config.api_key" type="password" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">JID (número destino)</label>
                <input v-model="channelForm.config.jid" class="form-input" placeholder="5511999999999" />
                <small class="hint">Formato E.164 sem "+" (ex: 5511999999999)</small>
              </div>
            </template>

            <!-- Webhook config -->
            <template v-if="channelForm.type === 'webhook'">
              <div class="form-group">
                <label class="form-label">URL</label>
                <input v-model="channelForm.config.url" class="form-input" placeholder="https://hooks.exemplo.com/..." />
              </div>
            </template>

            <div v-if="channelError" class="alert alert-danger" style="margin-bottom: 8px;">{{ channelError }}</div>
            <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 20px;">
              <button type="button" class="btn btn-secondary" @click="showAddChannel = false">Cancelar</button>
              <button type="submit" class="btn btn-primary" :disabled="savingChannel">
                {{ savingChannel ? 'Salvando...' : 'Criar Canal' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- SISTEMA -->
    <div v-if="currentTab === 'system'" class="card" style="max-width: 600px;">
      <div class="form-group">
        <label class="form-label">Retenção de Dados (dias)</label>
        <input v-model.number="system.retention_days" type="number" class="form-input" />
        <div class="form-hint">Por quantos dias manter o histórico de métricas.</div>
      </div>
      <div style="display: flex; align-items: center; gap: 12px;">
        <button class="btn btn-primary" :disabled="savingSystem" @click="saveSystem">
          {{ savingSystem ? 'Salvando...' : 'Salvar Sistema' }}
        </button>
        <span v-if="systemMsg" class="save-msg">{{ systemMsg }}</span>
      </div>
    </div>

    <!-- TEMPLATES -->
    <div v-if="currentTab === 'templates'">
      <SettingsTemplates />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../composables/useApi'
import { useAuthStore } from '../stores/auth'
import SettingsTemplates from '../components/SettingsTemplates.vue'

const auth = useAuthStore()

const tabs = [
  { id: 'profile', label: 'Perfil' },
  { id: 'channels', label: 'Canais de Alerta' },
  { id: 'templates', label: 'Templates' },
  { id: 'system', label: 'Sistema' },
]
const currentTab = ref('profile')

// ---- Profile ----
const profile = ref({ name: '', email: '', timezone: 'UTC', whatsapp_number: '' })
const savingProfile = ref(false)
const profileMsg = ref('')

function loadProfile() {
  if (auth.user) {
    profile.value = {
      name: auth.user.name || '',
      email: auth.user.email || '',
      timezone: auth.user.timezone || 'UTC',
      whatsapp_number: auth.user.whatsapp_number || '',
    }
  }
}

async function saveProfile() {
  savingProfile.value = true
  profileMsg.value = ''
  try {
    // Profile update via settings endpoint
    await api.put('/settings', profile.value)
    profileMsg.value = '✓ Salvo!'
    setTimeout(() => profileMsg.value = '', 3000)
  } catch (e) {
    profileMsg.value = 'Erro: ' + (e.response?.data?.error || e.message)
  } finally {
    savingProfile.value = false
  }
}

// ---- Channels ----
const channels = ref([])
const loadingChannels = ref(false)
const testResult = ref(null)
const showAddChannel = ref(false)
const savingChannel = ref(false)
const channelError = ref('')

function emptyChannelForm() {
  return { name: '', type: 'email', config: {} }
}
const channelForm = ref(emptyChannelForm())

async function loadChannels() {
  loadingChannels.value = true
  try {
    const res = await api.get('/settings/channels')
    channels.value = Array.isArray(res.data) ? res.data : []
  } catch {
    channels.value = []
  } finally {
    loadingChannels.value = false
  }
}

async function createChannel() {
  savingChannel.value = true
  channelError.value = ''
  try {
    const payload = {
      name: channelForm.value.name,
      type: channelForm.value.type,
      config: JSON.stringify(channelForm.value.config),
      enabled: true,
    }
    await api.post('/settings/channels', payload)
    showAddChannel.value = false
    channelForm.value = emptyChannelForm()
    await loadChannels()
  } catch (e) {
    channelError.value = e.response?.data?.error || e.message
  } finally {
    savingChannel.value = false
  }
}

async function updateChannel(ch) {
  try {
    await api.put(`/settings/channels/${ch.id}`, { enabled: ch.enabled })
  } catch { /* noop */ }
}

async function testChannel(ch) {
  testResult.value = null
  try {
    await api.post(`/settings/channels/${ch.id}/test`)
    testResult.value = { ok: true, message: `Teste de ${ch.type} enviado!` }
  } catch (e) {
    testResult.value = { ok: false, message: 'Falha: ' + (e.response?.data?.error || e.message) }
  }
}

async function deleteChannel(ch) {
  if (!confirm(`Remover canal "${ch.name}"?`)) return
  try {
    await api.delete(`/settings/channels/${ch.id}`)
    await loadChannels()
  } catch { /* noop */ }
}

// ---- System ----
const system = ref({ retention_days: 90 })
const savingSystem = ref(false)
const systemMsg = ref('')

async function loadSystem() {
  try {
    const res = await api.get('/settings')
    if (res.data && typeof res.data === 'object') {
      system.value = { retention_days: res.data.retention_days || 90 }
    }
  } catch { /* noop */ }
}

async function saveSystem() {
  savingSystem.value = true
  systemMsg.value = ''
  try {
    await api.put('/settings', { retention_days: system.value.retention_days })
    systemMsg.value = '✓ Salvo!'
    setTimeout(() => systemMsg.value = '', 3000)
  } catch (e) {
    systemMsg.value = 'Erro: ' + (e.response?.data?.error || e.message)
  } finally {
    savingSystem.value = false
  }
}

onMounted(() => {
  loadProfile()
  loadChannels()
  loadSystem()
})
</script>

<style scoped>
.tabs-header {
  display: flex;
  border-bottom: 1px solid var(--border);
  margin-bottom: 24px;
  overflow-x: auto;
}
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

.toggle { position: relative; display: inline-block; width: 40px; height: 22px; }
.toggle input { display: none; }
.toggle-slider {
  position: absolute;
  inset: 0;
  background: var(--bg-base);
  border: 1px solid var(--border-bright);
  border-radius: 22px;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.toggle-slider::before {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  left: 2px;
  top: 2px;
  background: var(--text-secondary);
  border-radius: 50%;
  transition: all var(--transition-fast);
}
.toggle input:checked + .toggle-slider {
  background: var(--accent-glow);
  border-color: var(--accent);
}
.toggle input:checked + .toggle-slider::before {
  transform: translateX(18px);
  background: var(--accent);
}

.provider-note { color: var(--text-muted); font-size: 0.82rem; margin: 0 0 14px; }
.provider-note code { background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 3px; font-size: 0.85em; }
.hint { color: var(--text-muted); font-size: 0.75rem; margin-top: 2px; }
.save-msg { color: var(--accent); font-size: 0.85rem; }
.empty-state { text-align: center; padding: 48px 20px; color: var(--text-muted); font-size: 13px; }

.modal-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
}
.modal {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  width: 90%; max-width: 500px; padding: 24px;
}
.modal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.modal-header h3 { margin: 0; font-size: 1rem; }
.modal form { display: flex; flex-direction: column; gap: 12px; }
</style>
