<template>
  <div class="auth-page">
    <div class="auth-box">
      <div class="auth-logo">
        <img src="/logo-full.png" alt="P-mon" class="auth-logo-img" />
      </div>
      <p class="auth-subtitle">Crie sua conta</p>

      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label class="form-label">Nome</label>
          <input v-model="form.name" class="form-input" placeholder="Seu nome" required />
        </div>
        <div class="form-group">
          <label class="form-label">Empresa</label>
          <input v-model="form.company_name" class="form-input" placeholder="Acme Corp" />
          <div class="form-hint">Nome da sua empresa (usado no whitelabel)</div>
        </div>
        <div class="form-group">
          <label class="form-label">Email</label>
          <input v-model="form.email" type="email" class="form-input" placeholder="voce@exemplo.com" required />
        </div>
        <div class="form-group">
          <label class="form-label">Senha</label>
          <input v-model="form.password" type="password" class="form-input" placeholder="mínimo 8 caracteres" minlength="8" required />
        </div>
        <div v-if="error" class="alert alert-danger" style="margin-bottom: 16px;">{{ error }}</div>
        <button type="submit" class="btn btn-primary btn-lg" style="width: 100%;" :disabled="loading">
          {{ loading ? 'Criando...' : 'Criar Conta' }}
        </button>
      </form>

      <div class="auth-footer">
        Já tem conta? <RouterLink to="/login">Entrar</RouterLink>
      </div>
    </div>
    <div class="auth-bg-grid"></div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const form = ref({ name: '', email: '', password: '', company_name: '' })

async function handleRegister() {
  loading.value = true
  error.value = ''
  try {
    await auth.register({
      name: form.value.name,
      email: form.value.email,
      password: form.value.password,
      company_name: form.value.company_name,
    })
    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.error || 'Falha ao criar conta.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; position: relative; overflow: hidden; }
.auth-bg-grid { position: absolute; inset: 0; background-image: linear-gradient(var(--border) 1px, transparent 1px), linear-gradient(90deg, var(--border) 1px, transparent 1px); background-size: 40px 40px; opacity: 0.15; mask-image: radial-gradient(ellipse at center, black 20%, transparent 70%); z-index: 0; }
.auth-box { width: 100%; max-width: 380px; background: var(--bg-card); border: 1px solid var(--border-bright); border-radius: var(--radius-xl); padding: 40px; box-shadow: 0 24px 80px rgba(0,0,0,0.5); position: relative; z-index: 1; }
.auth-logo { display: flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.auth-logo-img { max-width: 240px; max-height: 90px; object-fit: contain; }
.auth-subtitle { text-align: center; color: var(--text-secondary); font-size: 13px; margin-bottom: 32px; }
.auth-footer { text-align: center; margin-top: 24px; font-size: 13px; color: var(--text-secondary); }
.auth-footer a { color: var(--accent); text-decoration: none; font-weight: 500; }
</style>
