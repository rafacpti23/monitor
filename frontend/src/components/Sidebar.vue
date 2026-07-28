<template>
  <aside class="sidebar">
    <!-- Logo (whitelabel > default) -->
    <div class="sidebar-logo">
      <img :src="displayLogo" alt="logo" class="logo-img" />
      <div class="logo-text">{{ systemName }}</div>
    </div>

    <!-- Nav -->
    <nav class="sidebar-nav">
      <div class="nav-section-label">Principal</div>
      <RouterLink to="/dashboard" class="nav-item" :class="{ active: route.path === '/dashboard' }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 13h1v4H3v-4zm4-4h1v8H7V9zm4-5h1v13h-1V4zm4 7h1v6h-1v-6zm4-3h1v9h-1V8z"/>
        </svg>
        Dashboard
      </RouterLink>
      <RouterLink to="/servers" class="nav-item" :class="{ active: route.path.startsWith('/servers') }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="2" y="4" width="16" height="4" rx="1"/>
          <rect x="2" y="10" width="16" height="4" rx="1"/>
          <circle cx="5" cy="6" r="1" fill="currentColor"/>
          <circle cx="5" cy="12" r="1" fill="currentColor"/>
        </svg>
        Servidores
        <span v-if="alertCounts.servers > 0" class="nav-badge">{{ alertCounts.servers }}</span>
      </RouterLink>
      <RouterLink to="/websites" class="nav-item" :class="{ active: route.path.startsWith('/websites') }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="10" cy="10" r="8"/>
          <path d="M10 2c2.5 2 4 5 4 8s-1.5 6-4 8c-2.5-2-4-5-4-8s1.5-6 4-8z"/>
          <path d="M2 10h16"/>
        </svg>
        Websites
        <span v-if="alertCounts.websites > 0" class="nav-badge">{{ alertCounts.websites }}</span>
      </RouterLink>
      <RouterLink to="/checks" class="nav-item" :class="{ active: route.path.startsWith('/checks') }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M4 10l4 4 8-8"/>
          <circle cx="10" cy="10" r="8"/>
        </svg>
        Checks
      </RouterLink>
      <RouterLink to="/papi" class="nav-item" :class="{ active: route.path.startsWith('/papi') }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M10 2a8 8 0 0 0-7 12l-1 4 4-1a8 8 0 1 0 4-15z"/>
          <path d="M7 9c0 3 2 5 5 5" stroke-width="1.2"/>
        </svg>
        WhatsApp PAPI
        <span v-if="alertCounts.papi > 0" class="nav-badge">{{ alertCounts.papi }}</span>
      </RouterLink>

      <div class="nav-section-label" style="margin-top: 16px;">Sistema</div>
      <RouterLink to="/incidents" class="nav-item" :class="{ active: route.path === '/incidents' }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M10 2L2 16h16L10 2z"/>
          <path d="M10 9v4M10 14v1"/>
        </svg>
        Incidentes
        <span v-if="activeIncidents > 0" class="nav-badge">{{ activeIncidents }}</span>
      </RouterLink>
      <RouterLink to="/company" class="nav-item" :class="{ active: route.path === '/company' }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="3" y="6" width="14" height="11" rx="1"/>
          <path d="M7 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2M7 10h6M7 13h6"/>
        </svg>
        Empresa
      </RouterLink>
      <RouterLink to="/api-docs" class="nav-item" :class="{ active: route.path === '/api-docs' }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M7 6L3 10l4 4M13 6l4 4-4 4M11 3l-2 14"/>
        </svg>
        API Docs
      </RouterLink>
      <RouterLink to="/settings" class="nav-item" :class="{ active: route.path === '/settings' }">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="10" cy="10" r="2"/>
          <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.22 4.22l1.42 1.42M14.36 14.36l1.42 1.42M4.22 15.78l1.42-1.42M14.36 5.64l1.42-1.42"/>
        </svg>
        Configurações
      </RouterLink>
    </nav>

    <!-- Footer -->
    <div class="sidebar-footer">
      <button class="logout-btn" @click="handleLogout">
        <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M7 4H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3M10 10h7M14 7l3 3-3 3"/>
        </svg>
        Sair
      </button>
      <div class="live-indicator">
        <div class="live-dot"></div>
        <span>REAL-TIME</span>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { computed, ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const alertCounts = ref({ servers: 0, websites: 0, papi: 0 })
const activeIncidents = ref(0)

const systemName = computed(() =>
  auth.company?.system_name || auth.company?.name || 'P-mon'
)
const displayLogo = computed(() => auth.company?.logo_url || '/logo-icon.png')

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.logo-img {
  max-height: 32px;
  max-width: 140px;
  object-fit: contain;
}
.logout-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  margin-bottom: 12px;
  background: none;
  border: 1px solid var(--border, #21262d);
  border-radius: var(--radius-md, 8px);
  color: var(--text-secondary, #c9d1d9);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.logout-btn:hover {
  background: rgba(244,67,54,0.1);
  border-color: rgba(244,67,54,0.3);
  color: var(--accent-red, #f44336);
}
.logout-btn .nav-icon {
  width: 18px;
  height: 18px;
}
</style>
