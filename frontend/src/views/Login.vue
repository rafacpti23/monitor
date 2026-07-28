<template>
  <div class="auth-page">
    <div class="auth-box">
      <div class="auth-logo">
        <img src="/logo-full.png" alt="P-mon" class="auth-logo-img" />
      </div>
      <p class="auth-subtitle">Monitoramento leve para suas VPS</p>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label class="form-label">Email</label>
          <input v-model="email" type="email" class="form-input" placeholder="voce@exemplo.com" required />
        </div>
        <div class="form-group">
          <label class="form-label">Senha</label>
          <input v-model="password" type="password" class="form-input" placeholder="••••••••" required />
        </div>
        <div v-if="error" class="alert alert-danger" style="margin-bottom: 16px;">{{ error }}</div>
        <button type="submit" class="btn btn-primary btn-lg" style="width: 100%;" :disabled="loading">
          {{ loading ? 'Entrando...' : 'Entrar' }}
        </button>
      </form>

      <div class="auth-footer">
        Não tem conta? <RouterLink to="/register">Criar conta</RouterLink>
      </div>
    </div>

    <div class="auth-bg-grid"></div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    const redirect = route.query.redirect || '/dashboard'
    router.push(redirect)
  } catch (e) {
    error.value = e.response?.data?.error || 'Email ou senha inválidos.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}
.auth-bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--border) 1px, transparent 1px),
    linear-gradient(90deg, var(--border) 1px, transparent 1px);
  background-size: 40px 40px;
  opacity: 0.15;
  mask-image: radial-gradient(ellipse at center, black 20%, transparent 70%);
  z-index: 0;
}
.auth-box {
  width: 100%;
  max-width: 380px;
  background: var(--bg-card);
  border: 1px solid var(--border-bright);
  border-radius: var(--radius-xl);
  padding: 40px;
  box-shadow: 0 24px 80px rgba(0,0,0,0.5);
  position: relative;
  z-index: 1;
}
.auth-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}
.auth-logo-img {
  max-width: 240px;
  max-height: 90px;
  object-fit: contain;
}
.auth-subtitle {
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
  margin-bottom: 32px;
}
.auth-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 13px;
  color: var(--text-secondary);
}
.auth-footer a { color: var(--accent); text-decoration: none; font-weight: 500; }
</style>
