import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const company = ref(null)
  const token = ref(localStorage.getItem('p-mon-token') || null)

  function setToken(newToken) {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('p-mon-token', newToken)
    } else {
      localStorage.removeItem('p-mon-token')
    }
  }

  function applyBranding() {
    // Apply whitelabel accent color + page title from the company.
    if (company.value) {
      if (company.value.accent_color) {
        document.documentElement.style.setProperty('--accent', company.value.accent_color)
      }
      const sysName = company.value.system_name || company.value.name
      if (sysName) document.title = sysName
    }
  }

  async function login(email, password) {
    const res = await api.post('/auth/login', { email, password })
    setToken(res.data.token)
    user.value = res.data.user
    company.value = res.data.company || null
    applyBranding()
  }

  async function register(payload) {
    const res = await api.post('/auth/register', payload)
    // Register does not return a token; log in right after.
    await login(payload.email, payload.password)
  }

  function logout() {
    setToken(null)
    user.value = null
    company.value = null
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const res = await api.get('/auth/me')
      // /auth/me now returns { user, company }
      user.value = res.data.user || res.data
      company.value = res.data.company || null
      applyBranding()
    } catch {
      logout()
    }
  }

  function setCompany(c) {
    company.value = c
    applyBranding()
  }

  return { user, company, token, setToken, login, register, logout, fetchMe, setCompany, applyBranding }
})
