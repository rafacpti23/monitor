<template>
  <div class="company-page">
    <!-- Branding Section -->
    <section class="section-card">
      <h2 class="section-title">Marca / Whitelabel</h2>
      <p class="section-desc">Configure o nome e logo que aparecem no sistema para seus usuários.</p>

      <form class="form-grid" @submit.prevent="saveBranding">
        <div class="form-group">
          <label>Nome da Empresa</label>
          <input v-model="brand.name" class="input" placeholder="Acme Corp" />
        </div>
        <div class="form-group">
          <label>Nome do Sistema (exibido no menu)</label>
          <input v-model="brand.system_name" class="input" placeholder="P-mon" />
        </div>
        <div class="form-group full-width">
          <label>URL do Logo (PNG ou SVG)</label>
          <input v-model="brand.logo_url" class="input" placeholder="https://..." />
          <div v-if="brand.logo_url" class="logo-preview">
            <img :src="brand.logo_url" alt="preview" @error="brand.logo_url = ''" />
          </div>
        </div>
        <div class="form-group">
          <label>Cor Accent</label>
          <div class="color-row">
            <input type="color" v-model="brand.accent_color" class="color-picker" />
            <input v-model="brand.accent_color" class="input" style="max-width: 120px" />
          </div>
        </div>
        <div class="form-actions full-width">
          <button type="submit" class="btn btn-primary" :disabled="saving">
            {{ saving ? 'Salvando...' : 'Salvar Marca' }}
          </button>
          <span v-if="brandMsg" class="save-msg">{{ brandMsg }}</span>
        </div>
      </form>
    </section>

    <!-- Internal Users Section -->
    <section class="section-card" style="margin-top: 24px;">
      <div class="section-header">
        <h2 class="section-title">Usuários Internos</h2>
        <button class="btn btn-primary btn-sm" @click="showAddUser = true">+ Novo Usuário</button>
      </div>

      <table class="data-table" v-if="users.length">
        <thead>
          <tr>
            <th>Nome</th>
            <th>Email</th>
            <th>Perfil</th>
            <th>WhatsApp</th>
            <th style="width: 120px">Ações</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.name }}</td>
            <td>{{ u.email }}</td>
            <td><span class="role-badge" :class="u.role">{{ u.role }}</span></td>
            <td>{{ u.whatsapp_number || '-' }}</td>
            <td>
              <div class="row-actions">
                <button class="btn btn-ghost btn-sm" @click="editUser(u)" title="Editar">✎</button>
                <button class="btn btn-ghost btn-sm btn-danger" @click="removeUser(u)" title="Remover"
                  :disabled="u.id === auth.user?.id">✕</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty-state">Nenhum usuário cadastrado ainda.</p>
    </section>

    <!-- Add / Edit User Modal -->
    <div v-if="showAddUser || editingUser" class="modal-overlay" @click.self="closeUserModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ editingUser ? 'Editar Usuário' : 'Novo Usuário' }}</h3>
          <button class="btn btn-ghost btn-sm" @click="closeUserModal">✕</button>
        </div>
        <form @submit.prevent="saveUser">
          <div class="form-group">
            <label>Nome</label>
            <input v-model="userForm.name" class="input" required />
          </div>
          <div class="form-group" v-if="!editingUser">
            <label>Email</label>
            <input v-model="userForm.email" type="email" class="input" required />
          </div>
          <div class="form-group">
            <label>{{ editingUser ? 'Nova Senha (em branco = mantém)' : 'Senha' }}</label>
            <input v-model="userForm.password" type="password" class="input" :required="!editingUser" minlength="8" />
          </div>
          <div class="form-group">
            <label>Perfil</label>
            <select v-model="userForm.role" class="input">
              <option value="admin">Admin</option>
              <option value="member">Membro</option>
              <option value="viewer">Visualizador</option>
            </select>
          </div>
          <div class="form-group">
            <label>WhatsApp (JID)</label>
            <input v-model="userForm.whatsapp_number" class="input" placeholder="5511999999999" />
          </div>
          <div class="form-actions">
            <button type="button" class="btn btn-ghost" @click="closeUserModal">Cancelar</button>
            <button type="submit" class="btn btn-primary" :disabled="savingUser">
              {{ savingUser ? 'Salvando...' : 'Salvar' }}
            </button>
          </div>
          <p v-if="userError" class="error-text">{{ userError }}</p>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../composables/useApi'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

// ---- Branding ----
const brand = ref({ name: '', system_name: '', logo_url: '', accent_color: '#00e676' })
const saving = ref(false)
const brandMsg = ref('')

async function loadBrand() {
  try {
    const res = await api.get('/company')
    // Defensive: only accept a real object payload, never a string (e.g. HTML fallback).
    if (res.data && typeof res.data === 'object' && !Array.isArray(res.data)) {
      brand.value = { ...brand.value, ...res.data }
    }
  } catch { /* no company yet */ }
}

async function saveBranding() {
  saving.value = true
  brandMsg.value = ''
  try {
    const res = await api.put('/company', brand.value)
    brand.value = res.data
    auth.setCompany(res.data)
    brandMsg.value = '✓ Salvo!'
    setTimeout(() => brandMsg.value = '', 3000)
  } catch (e) {
    brandMsg.value = 'Erro: ' + (e.response?.data?.error || e.message)
  } finally {
    saving.value = false
  }
}

// ---- Users ----
const users = ref([])
const showAddUser = ref(false)
const editingUser = ref(null)
const savingUser = ref(false)
const userError = ref('')
const userForm = ref({ name: '', email: '', password: '', role: 'member', whatsapp_number: '' })

async function loadUsers() {
  try {
    const res = await api.get('/company/users')
    // Defensive: only accept a real array; guards against HTML SPA-fallback strings
    // being iterated char-by-char by v-for.
    users.value = Array.isArray(res.data) ? res.data : []
  } catch { users.value = [] }
}

function editUser(u) {
  editingUser.value = u
  userForm.value = { name: u.name, email: u.email, password: '', role: u.role, whatsapp_number: u.whatsapp_number || '' }
}

function closeUserModal() {
  showAddUser.value = false
  editingUser.value = null
  userError.value = ''
  userForm.value = { name: '', email: '', password: '', role: 'member', whatsapp_number: '' }
}

async function saveUser() {
  savingUser.value = true
  userError.value = ''
  try {
    if (editingUser.value) {
      await api.put('/company/users/' + editingUser.value.id, userForm.value)
    } else {
      await api.post('/company/users', userForm.value)
    }
    closeUserModal()
    await loadUsers()
  } catch (e) {
    userError.value = e.response?.data?.error || e.message
  } finally {
    savingUser.value = false
  }
}

async function removeUser(u) {
  if (!confirm(`Remover ${u.name} (${u.email})?`)) return
  try {
    await api.delete('/company/users/' + u.id)
    await loadUsers()
  } catch { /* */ }
}

onMounted(() => { loadBrand(); loadUsers() })
</script>

<style scoped>
.company-page { padding: 0; }
.section-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}
.section-title { margin: 0 0 4px; font-size: 1.1rem; color: var(--text-primary); }
.section-desc { margin: 0 0 20px; color: var(--text-muted); font-size: 0.85rem; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.full-width { grid-column: 1 / -1; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { color: var(--text-secondary); font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.5px; }

.color-row { display: flex; align-items: center; gap: 10px; }
.color-picker { width: 40px; height: 36px; border: 1px solid var(--border); border-radius: var(--radius); cursor: pointer; padding: 2px; background: transparent; }
.logo-preview { margin-top: 8px; }
.logo-preview img { max-height: 40px; max-width: 200px; object-fit: contain; }
.form-actions { display: flex; align-items: center; gap: 12px; margin-top: 8px; }
.save-msg { color: var(--accent); font-size: 0.85rem; }

.data-table { width: 100%; border-collapse: collapse; }
.data-table th,
.data-table td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); }
.data-table th { color: var(--text-muted); font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.5px; }
.data-table td { font-size: 0.85rem; color: var(--text-primary); }

.role-badge {
  font-size: 0.7rem;
  text-transform: uppercase;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 600;
  letter-spacing: 0.5px;
}
.role-badge.admin { background: rgba(0, 230, 118, 0.15); color: var(--accent); }
.role-badge.member { background: rgba(66, 165, 245, 0.15); color: #42a5f5; }
.role-badge.viewer { background: rgba(255, 255, 255, 0.08); color: var(--text-muted); }

.row-actions { display: flex; gap: 6px; }
.btn-danger { color: var(--accent-red) !important; }
.btn-danger:hover { background: rgba(244, 67, 54, 0.1); }

.empty-state { color: var(--text-muted); text-align: center; padding: 32px 0; font-size: 0.85rem; }

.modal-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
}
.modal {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  width: 90%; max-width: 440px; padding: 24px;
}
.modal-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.modal-header h3 { margin: 0; font-size: 1rem; }
.modal form { display: flex; flex-direction: column; gap: 12px; }
.error-text { color: var(--accent-red); font-size: 0.82rem; margin: 4px 0 0; }
</style>
