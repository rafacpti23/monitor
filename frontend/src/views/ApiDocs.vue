<template>
  <div class="apidocs">
    <div class="doc-layout">
      <!-- Sidebar nav -->
      <nav class="doc-nav">
        <a v-for="g in groups" :key="g.id" :href="'#' + g.id"
           class="doc-nav-item" :class="{ active: activeGroup === g.id }"
           @click="activeGroup = g.id">
          {{ g.title }}
        </a>
      </nav>

      <!-- Content -->
      <div class="doc-content">
        <header class="doc-header">
          <h1>Documentação da API</h1>
          <p>Base URL: <code>{{ baseUrl }}</code></p>
          <p class="auth-note">
            Autenticação via <code>Authorization: Bearer &lt;token&gt;</code> (obtido no login).
            O agente usa a rota pública <code>/agent/:key</code> com a chave do servidor.
          </p>
        </header>

        <section v-for="g in groups" :key="g.id" :id="g.id" class="doc-section">
          <h2>{{ g.title }}</h2>
          <p v-if="g.desc" class="group-desc">{{ g.desc }}</p>

          <div v-for="(ep, i) in g.endpoints" :key="i" class="endpoint">
            <div class="ep-head">
              <span class="method" :class="ep.method.toLowerCase()">{{ ep.method }}</span>
              <code class="ep-path">{{ ep.path }}</code>
            </div>
            <p class="ep-desc">{{ ep.desc }}</p>
            <pre v-if="ep.body" class="ep-body"><code>{{ ep.body }}</code></pre>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const baseUrl = import.meta.env.VITE_API_URL || '/api/v1'
const activeGroup = ref('auth')

const groups = [
  {
    id: 'auth', title: 'Autenticação',
    desc: 'Registro, login e sessão do usuário.',
    endpoints: [
      { method: 'POST', path: '/auth/register', desc: 'Cria uma empresa e um usuário admin.',
        body: '{\n  "email": "you@empresa.com",\n  "name": "Seu Nome",\n  "password": "min 8 chars",\n  "company_name": "Acme Corp"\n}' },
      { method: 'POST', path: '/auth/login', desc: 'Retorna token JWT + dados do usuário + empresa (whitelabel).',
        body: '{\n  "email": "you@empresa.com",\n  "password": "..."\n}' },
      { method: 'GET', path: '/auth/me', desc: 'Retorna o usuário atual e a empresa. Requer token.' },
      { method: 'POST', path: '/auth/logout', desc: 'Encerra a sessão.' },
    ],
  },
  {
    id: 'company', title: 'Empresa (Whitelabel)',
    desc: 'Marca personalizada e gestão de usuários internos. Escrita exige perfil admin.',
    endpoints: [
      { method: 'GET', path: '/company', desc: 'Retorna a empresa (nome, system_name, logo_url, accent_color).' },
      { method: 'PUT', path: '/company', desc: 'Atualiza marca. Admin.',
        body: '{\n  "name": "Acme Corp",\n  "system_name": "Acme Monitor",\n  "logo_url": "https://.../logo.png",\n  "accent_color": "#00e676"\n}' },
      { method: 'GET', path: '/company/users', desc: 'Lista usuários internos da empresa.' },
      { method: 'POST', path: '/company/users', desc: 'Cria usuário interno. Admin.',
        body: '{\n  "email": "user@empresa.com",\n  "name": "Nome",\n  "password": "min 8",\n  "role": "member",\n  "whatsapp_number": "5511999999999"\n}' },
      { method: 'PUT', path: '/company/users/:id', desc: 'Edita usuário. Senha em branco = mantém. Admin.',
        body: '{\n  "name": "Novo Nome",\n  "role": "viewer",\n  "whatsapp_number": "...",\n  "password": ""\n}' },
      { method: 'DELETE', path: '/company/users/:id', desc: 'Remove usuário (não pode ser você mesmo). Admin.' },
    ],
  },
  {
    id: 'servers', title: 'Servidores',
    endpoints: [
      { method: 'GET', path: '/servers', desc: 'Lista servidores com métricas mais recentes.' },
      { method: 'POST', path: '/servers', desc: 'Cria servidor e gera server_key.',
        body: '{\n  "name": "vps-prod-01",\n  "type": "linux"\n}' },
      { method: 'GET', path: '/servers/:id', desc: 'Detalhe do servidor + serviços.' },
      { method: 'PUT', path: '/servers/:id', desc: 'Atualiza servidor.' },
      { method: 'DELETE', path: '/servers/:id', desc: 'Remove servidor.' },
      { method: 'GET', path: '/servers/:id/history', desc: 'Histórico de métricas. Query: ?range=1h|6h|24h|7d|30d' },
      { method: 'GET', path: '/servers/:id/incidents', desc: 'Incidentes do servidor.' },
    ],
  },
  {
    id: 'agent', title: 'Agente (Receiver)',
    desc: 'Endpoint público usado pelo binário agente instalado no VPS. Aceita JSON ou JSON gzipado.',
    endpoints: [
      { method: 'POST', path: '/agent/:key', desc: 'Recebe payload de métricas. A key é o server_key do servidor.',
        body: '{\n  "hostname": "vps-prod-01",\n  "os": "Ubuntu 24.04",\n  "cpu_percent": 34.7,\n  "ram_total_bytes": 1073741824,\n  "ram_used_bytes": 536870912,\n  "load_avg": [0.42, 0.38, 0.35],\n  "disks": [{"path":"/","total_bytes":..,"used_bytes":..}],\n  "docker_containers": [...],\n  "pm2_processes": [...],\n  "services": [{"name":"nginx","running":true}]\n}' },
    ],
  },
  {
    id: 'websites', title: 'Websites',
    endpoints: [
      { method: 'GET', path: '/websites', desc: 'Lista websites monitorados.' },
      { method: 'POST', path: '/websites', desc: 'Adiciona website.',
        body: '{\n  "name": "Meu Site",\n  "url": "https://exemplo.com",\n  "method": "GET",\n  "check_interval_sec": 60,\n  "search_string": ""\n}' },
      { method: 'GET', path: '/websites/:id', desc: 'Detalhe do website.' },
      { method: 'PUT', path: '/websites/:id', desc: 'Atualiza website.' },
      { method: 'DELETE', path: '/websites/:id', desc: 'Remove website.' },
      { method: 'GET', path: '/websites/:id/history', desc: 'Histórico de respostas.' },
    ],
  },
  {
    id: 'checks', title: 'Checks',
    desc: 'Verificações de rede: ping, tcp, http, dns, ssl_expiry.',
    endpoints: [
      { method: 'GET', path: '/checks', desc: 'Lista checks.' },
      { method: 'POST', path: '/checks', desc: 'Cria check.',
        body: '{\n  "name": "DNS Google",\n  "type": "dns",\n  "target": "google.com",\n  "port": 0,\n  "interval_sec": 60\n}' },
      { method: 'GET', path: '/checks/:id', desc: 'Detalhe do check.' },
      { method: 'PUT', path: '/checks/:id', desc: 'Atualiza check.' },
      { method: 'DELETE', path: '/checks/:id', desc: 'Remove check.' },
    ],
  },
  {
    id: 'incidents', title: 'Incidentes',
    endpoints: [
      { method: 'GET', path: '/incidents', desc: 'Lista incidentes.' },
      { method: 'PUT', path: '/incidents/:id/acknowledge', desc: 'Reconhece incidente.' },
      { method: 'PUT', path: '/incidents/:id/resolve', desc: 'Resolve incidente.' },
      { method: 'PUT', path: '/incidents/:id/ignore', desc: 'Ignora incidente.' },
    ],
  },
  {
    id: 'alerts', title: 'Regras & Canais de Alerta',
    desc: 'Canais suportados: email, whatsapp (PAPI), webhook.',
    endpoints: [
      { method: 'GET', path: '/alert-rules', desc: 'Lista regras de alerta.' },
      { method: 'POST', path: '/alert-rules', desc: 'Cria regra.',
        body: '{\n  "monitor_type": "server",\n  "monitor_id": 1,\n  "alert_type": "cpu",\n  "comparison": ">=",\n  "threshold": "90",\n  "occurrences": 3,\n  "cooldown_min": 30,\n  "channels": "[1]"\n}' },
      { method: 'PUT', path: '/alert-rules/:id', desc: 'Atualiza regra.' },
      { method: 'DELETE', path: '/alert-rules/:id', desc: 'Remove regra.' },
      { method: 'GET', path: '/settings/channels', desc: 'Lista canais de notificação.' },
      { method: 'POST', path: '/settings/channels', desc: 'Cria canal. Para WhatsApp use type=whatsapp e config PAPI.',
        body: '// WhatsApp (PAPI)\n{\n  "type": "whatsapp",\n  "name": "Alertas WhatsApp",\n  "config": "{\\"instance\\":\\"minha_instancia\\",\\"api_key\\":\\"SUA_API_KEY\\",\\"jid\\":\\"5511999999999\\"}",\n  "enabled": true\n}' },
      { method: 'PUT', path: '/settings/channels/:id', desc: 'Atualiza canal.' },
      { method: 'DELETE', path: '/settings/channels/:id', desc: 'Remove canal.' },
      { method: 'POST', path: '/settings/channels/:id/test', desc: 'Envia mensagem de teste pelo canal.' },
    ],
  },
  {
    id: 'papi', title: 'WhatsApp via PAPI',
    desc: 'O único provedor de WhatsApp suportado. Você configura apenas: Instância, API Key e JID (número que recebe o alerta).',
    endpoints: [
      { method: 'POST', path: 'https://api.papi.api.br/api/instances/{instancia}/send-text',
        desc: 'Chamada real feita pelo backend ao disparar um alerta. Header x-api-key, corpo jid + text.',
        body: 'curl -X POST "https://api.papi.api.br/api/instances/{instancia}/send-text" \\\n  -H "x-api-key: SUA_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d \'{\n    "jid": "5511999999999",\n    "text": "Olá do Papi!"\n  }\'' },
    ],
  },
]
</script>

<style scoped>
.apidocs { padding: 0; }
.doc-layout { display: grid; grid-template-columns: 200px 1fr; gap: 24px; align-items: start; }

.doc-nav {
  position: sticky; top: 16px;
  display: flex; flex-direction: column; gap: 2px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 12px;
}
.doc-nav-item {
  padding: 8px 12px;
  border-radius: var(--radius);
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.85rem;
  transition: all 0.15s;
}
.doc-nav-item:hover { background: var(--surface-hover); color: var(--text-primary); }
.doc-nav-item.active { background: rgba(0,230,118,0.12); color: var(--accent); }

.doc-content { min-width: 0; }
.doc-header {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 24px; margin-bottom: 20px;
}
.doc-header h1 { margin: 0 0 8px; font-size: 1.4rem; color: var(--text-primary); }
.doc-header p { margin: 4px 0; color: var(--text-muted); font-size: 0.85rem; }
.auth-note { margin-top: 12px !important; }

.doc-section {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 24px; margin-bottom: 20px;
}
.doc-section h2 { margin: 0 0 4px; font-size: 1.1rem; color: var(--accent); }
.group-desc { margin: 0 0 16px; color: var(--text-muted); font-size: 0.85rem; }

.endpoint {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 14px 16px; margin-bottom: 12px;
}
.ep-head { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.method {
  font-family: var(--font-mono, monospace);
  font-size: 0.72rem; font-weight: 700;
  padding: 3px 8px; border-radius: 4px;
  min-width: 52px; text-align: center;
}
.method.get { background: rgba(66,165,245,0.15); color: #42a5f5; }
.method.post { background: rgba(0,230,118,0.15); color: var(--accent); }
.method.put { background: rgba(255,193,7,0.15); color: #ffc107; }
.method.delete { background: rgba(244,67,54,0.15); color: var(--accent-red); }
.ep-path { font-family: var(--font-mono, monospace); font-size: 0.85rem; color: var(--text-primary); word-break: break-all; }
.ep-desc { margin: 8px 0 0; color: var(--text-secondary); font-size: 0.82rem; }
.ep-body {
  margin: 10px 0 0; padding: 12px;
  background: var(--bg, #0a0e12);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow-x: auto;
}
.ep-body code { font-family: var(--font-mono, monospace); font-size: 0.78rem; color: var(--text-secondary); white-space: pre; }

code { font-family: var(--font-mono, monospace); background: rgba(255,255,255,0.06); padding: 1px 5px; border-radius: 3px; font-size: 0.85em; }

@media (max-width: 768px) {
  .doc-layout { grid-template-columns: 1fr; }
  .doc-nav { position: static; flex-direction: row; flex-wrap: wrap; }
}
</style>
