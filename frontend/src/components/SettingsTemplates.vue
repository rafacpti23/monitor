<template>
  <div class="settings-section">
    <div class="section-header">
      <div>
        <h2 class="section-title">Templates de Mensagens</h2>
        <p class="section-desc">Personalize as notificações para cada tipo de alerta. Use as variáveis disponíveis para incluir informações dinâmicas.</p>
      </div>
      <div class="section-badge">
        <span class="badge-count">{{ templates.filter(t => t.id).length }}</span>
        <span class="badge-label">customizados</span>
      </div>
    </div>

    <div class="templates-grid">
      <div v-for="tpl in templates" :key="tpl.alert_type" class="template-card" :class="{ customized: tpl.id }">
        <!-- Card Header -->
        <div class="card-header">
          <div class="alert-type-icon" :class="`icon-${tpl.alert_type}`">
            <span class="icon-emoji">{{ getEmoji(tpl.alert_type) }}</span>
          </div>
          <div class="header-text">
            <div class="type-label">{{ formatAlertType(tpl.alert_type) }}</div>
            <div class="status-row">
              <span class="status-dot" :class="tpl.id ? 'custom' : 'default'"></span>
              <span class="status-text">{{ tpl.id ? 'Personalizado' : 'Padrão do sistema' }}</span>
            </div>
          </div>
          <button v-if="tpl.id" class="reset-btn" @click="resetTemplate(tpl)" title="Restaurar padrão">
            <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 10a7 7 0 0 1 13.5-2.5M17 10a7 7 0 0 1-13.5 2.5"/>
              <path d="M3 3v5h5M17 17v-5h-5"/>
            </svg>
          </button>
        </div>

        <!-- Form -->
        <div class="card-form">
          <div class="form-field">
            <label class="field-label">Assunto</label>
            <input
              v-model="tpl.subject"
              type="text"
              class="field-input"
              :placeholder="'Assunto da notificação...'"
            />
            <div class="field-counter">{{ tpl.subject.length }} caracteres</div>
          </div>

          <div class="form-field">
            <label class="field-label">Mensagem</label>
            <textarea
              v-model="tpl.body"
              class="field-textarea"
              rows="6"
              :placeholder="'Conteúdo da mensagem...'"
            ></textarea>
            <div class="field-counter">{{ tpl.body.length }} caracteres</div>
          </div>

          <!-- Variables -->
          <div class="vars-section">
            <div class="vars-title">Variáveis disponíveis (clique para copiar)</div>
            <div class="vars-grid">
              <button
                v-for="v in getVariables(tpl.alert_type)"
                :key="v"
                class="var-chip"
                @click="copyVar(v)"
              >
                {{ v }}
              </button>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="card-footer">
          <div class="save-status" :class="{ success: tpl._saved, error: tpl._error }">
            <span v-if="tpl._saved">✓ Salvo!</span>
            <span v-else-if="tpl._error">✗ Erro ao salvar</span>
          </div>
          <button
            class="btn btn-primary"
            :disabled="tpl._saving"
            @click="saveTemplate(tpl)"
          >
            {{ tpl._saving ? 'Salvando...' : 'Salvar' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../composables/useApi'

const defaultTemplates = [
  {
    alert_type: 'website_down',
    subject: '🌐❌ Site fora do ar: {{monitor_name}}',
    body: '🌐❌ *SITE FORA DO AR*\n\n📛 URL: {{monitor_url}}\n⏱️ Detectado às: {{timestamp}}\n🔴 Status: Indisponível'
  },
  {
    alert_type: 'server_offline',
    subject: '🖥️💀 Servidor offline: {{monitor_name}}',
    body: '🖥️💀 *SERVIDOR OFFLINE*\n\n📛 Servidor: *{{monitor_name}}*\n🕐 Hora: {{timestamp}}\n🔴 Status: Sem resposta'
  },
  {
    alert_type: 'check_failed',
    subject: '⚠️ Verificação falhou: {{monitor_name}}',
    body: '⚠️ *VERIFICAÇÃO FALHOU*\n\n📛 Monitor: *{{monitor_name}}*\n⚠️ Erro: {{error_message}}\n🕐 Hora: {{timestamp}}'
  },
  {
    alert_type: 'no_data',
    subject: '📡 Sem dados: {{monitor_name}}',
    body: '📡 *SEM DADOS DO AGENTE*\n\n📛 Servidor: *{{monitor_name}}*\n⏱️ Sem envio há: {{duration}}\n🕐 Hora: {{timestamp}}'
  },
  {
    alert_type: 'resolved',
    subject: '✅ Normalizado: {{monitor_name}}',
    body: '✅ *VOLTOU AO NORMAL*\n\n📛 Monitor: *{{monitor_name}}*\n⏱️ Duração do incidente: {{duration}}\n🕐 Hora: {{timestamp}}'
  }
]

const templates = ref(defaultTemplates.map(t => ({ ...t, _saving: false, _saved: false, _error: false })))

onMounted(loadTemplates)

async function loadTemplates() {
  try {
    const res = await api.get('/alert-templates')
    const custom = res.data || []
    templates.value.forEach(tpl => {
      const c = custom.find(x => x.alert_type === tpl.alert_type)
      if (c) {
        tpl.id = c.id
        tpl.subject = c.subject
        tpl.body = c.body
      }
    })
  } catch (e) {
    console.error('Failed to load templates', e)
  }
}

async function saveTemplate(tpl) {
  tpl._saving = true
  tpl._saved = false
  tpl._error = false
  try {
    if (tpl.id) {
      await api.put(`/alert-templates/${tpl.id}`, { subject: tpl.subject, body: tpl.body })
    } else {
      const res = await api.post('/alert-templates', {
        alert_type: tpl.alert_type,
        subject: tpl.subject,
        body: tpl.body
      })
      tpl.id = res.data.id
    }
    tpl._saved = true
    setTimeout(() => { tpl._saved = false }, 3000)
  } catch (e) {
    tpl._error = true
    setTimeout(() => { tpl._error = false }, 4000)
    console.error('Failed to save template', e)
  } finally {
    tpl._saving = false
  }
}

async function resetTemplate(tpl) {
  if (!confirm('Restaurar para o padrão do sistema?')) return
  try {
    await api.post(`/alert-templates/${tpl.alert_type}/reset`)
    const def = defaultTemplates.find(d => d.alert_type === tpl.alert_type)
    if (def) {
      tpl.id = null
      tpl.subject = def.subject
      tpl.body = def.body
    }
  } catch (e) {
    console.error('Failed to reset template', e)
  }
}

function getEmoji(type) {
  const map = {
    website_down: '🌐',
    server_offline: '🖥️',
    check_failed: '⚠️',
    no_data: '📡',
    resolved: '✅'
  }
  return map[type] || '📋'
}

function formatAlertType(type) {
  const names = {
    website_down: 'Site Fora do Ar',
    server_offline: 'Servidor Offline',
    check_failed: 'Verificação Falhou',
    no_data: 'Sem Dados',
    resolved: 'Voltou ao Normal'
  }
  return names[type] || type
}

function getVariables(type) {
  const base = ['{{monitor_name}}', '{{timestamp}}']
  const extra = {
    website_down: ['{{monitor_url}}'],
    server_offline: [],
    check_failed: ['{{error_message}}'],
    no_data: ['{{duration}}'],
    resolved: ['{{duration}}']
  }
  return [...base, ...(extra[type] || [])]
}

async function copyVar(v) {
  try {
    await navigator.clipboard.writeText(v)
  } catch (e) {
    console.error('Copy failed', e)
  }
}
</script>

<style scoped>
.settings-section {
  margin-bottom: 48px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  gap: 20px;
}

.section-title {
  font-size: 22px;
  font-weight: 700;
  margin-bottom: 6px;
  color: var(--text-primary);
  letter-spacing: -0.4px;
}

.section-desc {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.section-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 14px 20px;
  background: linear-gradient(135deg, rgba(0, 230, 118, 0.1), rgba(0, 230, 118, 0.05));
  border-radius: 12px;
  border: 1px solid rgba(0, 230, 118, 0.2);
  box-shadow: 0 4px 12px rgba(0, 230, 118, 0.08);
}

.badge-count {
  font-size: 28px;
  font-weight: 700;
  color: var(--accent);
  font-family: 'JetBrains Mono', monospace;
  line-height: 1;
}

.badge-label {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  font-weight: 600;
}

.templates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(520px, 1fr));
  gap: 24px;
}

.template-card {
  background: linear-gradient(145deg, var(--bg-secondary), rgba(5, 15, 12, 0.6));
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
  overflow: hidden;
}

.template-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.template-card.customized::before {
  opacity: 1;
}

.template-card:hover {
  border-color: var(--border-bright);
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 230, 118, 0.12), 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* Card Header */
.card-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.alert-type-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  background: var(--bg-tertiary);
  border: 1.5px solid var(--border-color);
  transition: all 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.alert-type-icon.icon-website_down {
  background: linear-gradient(135deg, rgba(33, 150, 243, 0.2), rgba(33, 150, 243, 0.1));
  border-color: rgba(33, 150, 243, 0.4);
}

.alert-type-icon.icon-server_offline {
  background: linear-gradient(135deg, rgba(244, 67, 54, 0.2), rgba(244, 67, 54, 0.1));
  border-color: rgba(244, 67, 54, 0.4);
}

.alert-type-icon.icon-check_failed {
  background: linear-gradient(135deg, rgba(255, 152, 0, 0.2), rgba(255, 152, 0, 0.1));
  border-color: rgba(255, 152, 0, 0.4);
}

.alert-type-icon.icon-no_data {
  background: linear-gradient(135deg, rgba(156, 39, 176, 0.2), rgba(156, 39, 176, 0.1));
  border-color: rgba(156, 39, 176, 0.4);
}

.alert-type-icon.icon-resolved {
  background: linear-gradient(135deg, rgba(0, 230, 118, 0.2), rgba(0, 230, 118, 0.1));
  border-color: rgba(0, 230, 118, 0.4);
}

.template-card:hover .alert-type-icon {
  transform: scale(1.08) rotate(3deg);
}

.icon-emoji {
  display: block;
}

.header-text {
  flex: 1;
  min-width: 0;
}

.type-label {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  letter-spacing: -0.2px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.custom {
  background: var(--accent);
  box-shadow: 0 0 10px rgba(0, 230, 118, 0.6);
  animation: pulse-glow 2s ease-in-out infinite;
}

.status-dot.default {
  background: var(--text-muted);
}

@keyframes pulse-glow {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 10px rgba(0, 230, 118, 0.6);
  }
  50% {
    opacity: 0.7;
    box-shadow: 0 0 15px rgba(0, 230, 118, 0.8);
  }
}

.status-text {
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.6px;
  font-weight: 600;
}

.reset-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--bg-tertiary);
  border: 1.5px solid var(--border-color);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.reset-btn:hover {
  background: linear-gradient(135deg, rgba(255, 82, 82, 0.15), rgba(255, 82, 82, 0.1));
  border-color: rgba(255, 82, 82, 0.5);
  color: #ff5252;
  transform: rotate(180deg) scale(1.05);
}

/* Form */
.card-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.8px;
}

.field-input,
.field-textarea {
  width: 100%;
  padding: 12px 14px;
  background: var(--bg-tertiary);
  border: 1.5px solid var(--border-color);
  border-radius: 10px;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 14px;
  box-sizing: border-box;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.field-input:focus,
.field-textarea:focus {
  outline: none;
  border-color: var(--accent);
  background: rgba(0, 20, 10, 0.4);
  box-shadow: 0 0 0 4px rgba(0, 230, 118, 0.1), 0 4px 12px rgba(0, 230, 118, 0.15);
  transform: translateY(-1px);
}

.field-textarea {
  resize: vertical;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.65;
  min-height: 140px;
}

.field-counter {
  font-size: 11px;
  color: var(--text-muted);
  text-align: right;
  font-family: 'JetBrains Mono', monospace;
  margin-top: 2px;
}

/* Variables */
.vars-section {
  padding: 16px;
  background: rgba(0, 20, 10, 0.3);
  border: 1.5px solid var(--border-color);
  border-radius: 12px;
  transition: all 0.25s ease;
}

.vars-section:hover {
  background: rgba(0, 20, 10, 0.4);
  border-color: rgba(0, 230, 118, 0.15);
}

.vars-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.6px;
  margin-bottom: 12px;
}

.vars-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.var-chip {
  display: inline-flex;
  align-items: center;
  padding: 7px 12px;
  background: linear-gradient(135deg, rgba(0, 230, 118, 0.15), rgba(0, 230, 118, 0.1));
  border: 1.5px solid rgba(0, 230, 118, 0.3);
  border-radius: 8px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--accent);
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  user-select: none;
  box-shadow: 0 2px 6px rgba(0, 230, 118, 0.1);
}

.var-chip:hover {
  background: linear-gradient(135deg, rgba(0, 230, 118, 0.25), rgba(0, 230, 118, 0.15));
  border-color: rgba(0, 230, 118, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 230, 118, 0.2);
}

.var-chip:active {
  transform: translateY(0) scale(0.98);
}

/* Footer */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.save-status {
  font-size: 13px;
  font-weight: 600;
  font-family: 'JetBrains Mono', monospace;
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 20px;
  transition: all 0.3s ease;
}

.save-status.success {
  color: var(--accent);
  text-shadow: 0 0 8px rgba(0, 230, 118, 0.4);
}

.save-status.error {
  color: #ff5252;
  text-shadow: 0 0 8px rgba(255, 82, 82, 0.4);
}

/* Responsive */
@media (max-width: 1200px) {
  .templates-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    gap: 16px;
  }

  .template-card {
    padding: 20px;
  }

  .card-header {
    flex-wrap: wrap;
  }
}
</style>
