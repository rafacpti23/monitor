<template>
  <div style="max-width: 800px; margin: 0 auto;">
    <div class="page-header">
      <div>
        <h1 class="page-title">Adicionar Servidor</h1>
        <p class="page-subtitle">Configure o servidor e instale o agente na sua VPS</p>
      </div>
    </div>

    <!-- Step indicator -->
    <div class="steps">
      <div class="step" :class="{ active: step >= 1, done: step > 1 }">
        <div class="step-num">1</div>
        <span>Configurar</span>
      </div>
      <div class="step-line" :class="{ done: step > 1 }"></div>
      <div class="step" :class="{ active: step >= 2 }">
        <div class="step-num">2</div>
        <span>Instalar Agente</span>
      </div>
    </div>

    <!-- STEP 1: Config -->
    <div v-if="step === 1" class="card">
      <div class="form-group">
        <label class="form-label">Nome do Servidor</label>
        <input v-model="form.name" class="form-input" placeholder="ex: web-prod-01" />
      </div>

      <div class="form-group">
        <label class="form-label">Tipo</label>
        <select v-model="form.type" class="form-select">
          <option value="linux">Linux</option>
          <option value="windows">Windows</option>
          <option value="macos">macOS</option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">Intervalo de Envio (agente)</label>
        <select v-model="form.interval" class="form-select">
          <option value="30">30 segundos</option>
          <option value="60">1 minuto</option>
          <option value="180">3 minutos (recomendado)</option>
          <option value="300">5 minutos</option>
        </select>
        <div class="form-hint">O alerta de "offline" dispara em 2× este intervalo sem dados</div>
      </div>

      <!-- What to monitor -->
      <div class="form-group">
        <label class="form-label">O que o agente deve coletar</label>
        <div class="monitor-options">
          <label class="monitor-opt" :class="{ checked: form.collect.system }">
            <input type="checkbox" v-model="form.collect.system" />
            <div class="opt-icon">📊</div>
            <div class="opt-info">
              <div class="opt-title">Sistema</div>
              <div class="opt-desc">CPU, RAM, Disco, Load, Rede, Uptime</div>
            </div>
            <div class="opt-check"></div>
          </label>

          <label class="monitor-opt" :class="{ checked: form.collect.docker }">
            <input type="checkbox" v-model="form.collect.docker" />
            <div class="opt-icon">🐳</div>
            <div class="opt-info">
              <div class="opt-title">Docker</div>
              <div class="opt-desc">Containers rodando, alerta se algum cair</div>
            </div>
            <div class="opt-check"></div>
          </label>

          <label class="monitor-opt" :class="{ checked: form.collect.pm2 }">
            <input type="checkbox" v-model="form.collect.pm2" />
            <div class="opt-icon">⚙️</div>
            <div class="opt-info">
              <div class="opt-title">PM2</div>
              <div class="opt-desc">Processos Node.js, status online/offline</div>
            </div>
            <div class="opt-check"></div>
          </label>

          <label class="monitor-opt" :class="{ checked: form.collect.services }">
            <input type="checkbox" v-model="form.collect.services" />
            <div class="opt-icon">🔧</div>
            <div class="opt-info">
              <div class="opt-title">Serviços do Sistema</div>
              <div class="opt-desc">nginx, mysql, redis, etc. via systemctl</div>
            </div>
            <div class="opt-check"></div>
          </label>
        </div>
      </div>

      <!-- Services list (if enabled) -->
      <div v-if="form.collect.services" class="form-group">
        <label class="form-label">Serviços a monitorar (separados por vírgula)</label>
        <input v-model="form.servicesList" class="form-input" placeholder="nginx, mysql, redis" />
      </div>

      <!-- Map -->
      <div class="form-group">
        <label class="monitor-opt" :class="{ checked: form.onMap }" style="max-width: 100%;">
          <input type="checkbox" v-model="form.onMap" />
          <div class="opt-icon">🗺️</div>
          <div class="opt-info">
            <div class="opt-title">Mostrar no Mapa</div>
            <div class="opt-desc">Exibir localização geográfica no dashboard</div>
          </div>
          <div class="opt-check"></div>
        </label>
      </div>

      <div v-if="error" class="alert alert-danger" style="margin-bottom: 16px;">{{ error }}</div>

      <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
        <button class="btn btn-secondary" @click="$router.push('/servers')">Cancelar</button>
        <button class="btn btn-primary" @click="createServer" :disabled="!form.name || loading">
          {{ loading ? 'Criando...' : 'Criar & Gerar Agente' }}
        </button>
      </div>
    </div>

    <!-- STEP 2: Install -->
    <div v-if="step === 2">
      <div class="alert alert-success" style="margin-bottom: 20px;">
        <div>
          <strong>Servidor criado!</strong> Agora instale o agente na sua VPS com o comando abaixo.
        </div>
      </div>

      <div class="card" style="margin-bottom: 16px;">
        <div class="card-header">
          <h3 style="font-size: 13px;">Instalação em 1 linha</h3>
          <span class="badge badge-info">Linux</span>
        </div>
        <div class="code-block">
          <code>{{ installCommand }}</code>
          <button class="copy-btn" @click="copy(installCommand, 'install')">
            {{ copied === 'install' ? '✓ Copiado' : 'Copiar' }}
          </button>
        </div>
        <div class="form-hint" style="margin-top: 8px;">
          Cole isso no terminal da sua VPS. O agente será instalado como serviço e começará a enviar dados automaticamente.
        </div>
      </div>

      <div class="card" style="margin-bottom: 16px;">
        <div class="card-header">
          <h3 style="font-size: 13px;">Chave do Servidor (server_key)</h3>
        </div>
        <div class="code-block">
          <code>{{ serverKey }}</code>
          <button class="copy-btn" @click="copy(serverKey, 'key')">
            {{ copied === 'key' ? '✓ Copiado' : 'Copiar' }}
          </button>
        </div>
      </div>

      <div class="card" style="margin-bottom: 16px;">
        <div class="card-header">
          <h3 style="font-size: 13px;">Config manual (/etc/p-mon-agent.json)</h3>
        </div>
        <div class="code-block" style="white-space: pre;"><code>{{ configJson }}</code>
          <button class="copy-btn" @click="copy(configJson, 'config')">
            {{ copied === 'config' ? '✓ Copiado' : 'Copiar' }}
          </button>
        </div>
      </div>

      <!-- Waiting for first data -->
      <div class="card waiting-card">
        <div class="waiting-spinner"></div>
        <div>
          <div style="font-weight: 500;">Aguardando primeiro sinal do agente...</div>
          <div style="font-size: 12px; color: var(--text-muted); font-family: var(--font-mono);">
            Assim que o agente enviar dados, o servidor ficará online automaticamente.
          </div>
        </div>
      </div>

      <div style="display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;">
        <button class="btn btn-secondary" @click="$router.push('/servers')">Ver Servidores</button>
        <button class="btn btn-primary" @click="$router.push('/dashboard')">Ir para Dashboard</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import api from '../composables/useApi'

const step = ref(1)
const copied = ref('')
const serverKey = ref('')
const loading = ref(false)
const error = ref('')

const form = ref({
  name: '',
  type: 'linux',
  interval: '180',
  onMap: false,
  collect: {
    system: true,
    docker: false,
    pm2: false,
    services: false,
  },
  servicesList: 'nginx',
})

const backendUrl = computed(() => {
  const raw = import.meta.env.VITE_API_URL || window.location.origin
  return raw.replace(/\/api\/v1\/?$/, '')
})

const installCommand = computed(() => {
  const c = form.value.collect
  const params = new URLSearchParams()
  params.set('key', serverKey.value)
  params.set('interval', String(parseInt(form.value.interval) || 180))
  params.set('system', c.system ? '1' : '0')
  params.set('docker', c.docker ? '1' : '0')
  params.set('pm2', c.pm2 ? '1' : '0')
  if (c.services) {
    const svc = form.value.servicesList.split(',').map(s => s.trim()).filter(Boolean).join(',')
    if (svc) params.set('services', svc)
  }
  return `curl -sSL "${backendUrl.value}/install/agent.sh?${params.toString()}" | sudo bash`
})

const configJson = computed(() => {
  const services = form.value.collect.services
    ? form.value.servicesList.split(',').map(s => s.trim()).filter(Boolean)
    : []
  return JSON.stringify({
    backend_url: backendUrl.value,
    server_key: serverKey.value,
    interval_seconds: parseInt(form.value.interval),
    collect: {
      system: form.value.collect.system,
      docker: form.value.collect.docker,
      pm2: form.value.collect.pm2,
      services,
    },
  }, null, 2)
})

async function createServer() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.post('/servers', {
      name: form.value.name,
      type: form.value.type,
      on_map: form.value.onMap,
    })
    serverKey.value = res.data.server_key || res.data.key || ''
    if (!serverKey.value) {
      throw new Error('Backend não retornou server_key')
    }
    step.value = 2
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Falha ao criar servidor.'
  } finally {
    loading.value = false
  }
}

async function copy(text, which) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = which
    setTimeout(() => { copied.value = '' }, 2000)
  } catch (e) {
    console.error(e)
  }
}
</script>

<style scoped>
.steps { display: flex; align-items: center; justify-content: center; gap: 12px; margin-bottom: 32px; }
.step { display: flex; align-items: center; gap: 10px; color: var(--text-muted); font-size: 13px; font-weight: 500; }
.step.active { color: var(--text-primary); }
.step-num {
  width: 28px; height: 28px; border-radius: 50%;
  border: 2px solid var(--border-bright);
  display: flex; align-items: center; justify-content: center;
  font-family: var(--font-mono); font-size: 12px; font-weight: 700;
}
.step.active .step-num { border-color: var(--accent); background: var(--accent-glow); color: var(--accent); }
.step.done .step-num { background: var(--accent); border-color: var(--accent); color: var(--bg-base); }
.step-line { width: 80px; height: 2px; background: var(--border-bright); }
.step-line.done { background: var(--accent); }

.monitor-options { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.monitor-opt {
  display: flex; align-items: center; gap: 12px;
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  position: relative;
}
.monitor-opt:hover { border-color: var(--border-bright); }
.monitor-opt.checked { border-color: var(--accent); background: var(--accent-glow); }
.monitor-opt input { display: none; }
.opt-icon { font-size: 22px; }
.opt-info { flex: 1; }
.opt-title { font-weight: 600; font-size: 13px; }
.opt-desc { font-size: 11px; color: var(--text-muted); }
.opt-check { width: 18px; height: 18px; border: 2px solid var(--border-bright); border-radius: 50%; flex-shrink: 0; }
.monitor-opt.checked .opt-check { border-color: var(--accent); background: var(--accent); position: relative; }
.monitor-opt.checked .opt-check::after {
  content: '✓';
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  color: var(--bg-base); font-size: 11px; font-weight: bold;
}

.code-block {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 14px 16px;
  font-family: var(--font-mono);
  font-size: 12.5px;
  color: var(--accent);
  position: relative;
  overflow-x: auto;
  word-break: break-all;
}
.copy-btn {
  position: absolute; top: 8px; right: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border-bright);
  color: var(--text-secondary);
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  font-family: var(--font-body);
  transition: all var(--transition-fast);
}
.copy-btn:hover { color: var(--accent); border-color: var(--accent); }

.waiting-card { display: flex; align-items: center; gap: 16px; }
.waiting-spinner {
  width: 24px; height: 24px;
  border: 3px solid var(--border-bright);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
