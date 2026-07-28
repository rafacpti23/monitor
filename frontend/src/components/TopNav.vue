<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useIncidentsStore } from '@/stores/incidents'

const router = useRouter()
const authStore = useAuthStore()
const incidentsStore = useIncidentsStore()

const userName = computed(() => authStore.user?.name || 'User')
const activeAlerts = computed(() => incidentsStore.activeIncidents)

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <header class="topnav">
    <div class="topnav-left">
      <button class="mobile-menu-btn btn-ghost btn btn-icon" @click="$emit('toggle-sidebar')">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="3" y1="6" x2="21" y2="6"/>
          <line x1="3" y1="12" x2="21" y2="12"/>
          <line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
    </div>
    
    <div class="topnav-center">
      <div class="search-box">
        <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input 
          type="text" 
          class="search-input" 
          placeholder="Search servers, websites, checks..."
        />
        <kbd class="search-kbd">⌘K</kbd>
      </div>
    </div>
    
    <div class="topnav-right">
      <div class="connection-status" :class="{ connected: true }">
        <span class="status-dot online"></span>
        <span class="status-text">Live</span>
      </div>
      
      <button class="btn btn-ghost btn-icon notification-btn" @click="router.push('/incidents')">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
        <span v-if="activeAlerts > 0" class="notification-badge">{{ activeAlerts }}</span>
      </button>
      
      <div class="user-menu">
        <button class="user-btn">
          <div class="user-avatar">
            {{ userName.charAt(0).toUpperCase() }}
          </div>
          <span class="user-name">{{ userName }}</span>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </button>
        
        <div class="user-dropdown">
          <router-link to="/settings" class="dropdown-item">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            Settings
          </router-link>
          <button class="dropdown-item" @click="handleLogout">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16 17 21 12 16 7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
            Logout
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.topnav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 var(--space-lg);
  background-color: var(--bg-surface);
  border-bottom: 1px solid var(--border);
  gap: var(--space-lg);
}

.topnav-left {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.mobile-menu-btn {
  display: none;
}

@media (max-width: 768px) {
  .mobile-menu-btn {
    display: flex;
  }
}

.topnav-center {
  flex: 1;
  max-width: 480px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.search-box:focus-within {
  border-color: var(--accent-green);
  box-shadow: 0 0 0 3px var(--accent-green-glow);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-kbd {
  padding: 2px 6px;
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--text-muted);
  background-color: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}

.topnav-right {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.connection-status {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xs) var(--space-sm);
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.connection-status.connected .status-text {
  color: var(--accent-green);
}

.notification-btn {
  position: relative;
}

.notification-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: white;
  background-color: var(--accent-red);
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-menu {
  position: relative;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-xs) var(--space-sm);
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.user-btn:hover {
  background-color: var(--bg-card);
  border-color: var(--border);
}

.user-avatar {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--bg-base);
  background-color: var(--accent-green);
  border-radius: 50%;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 500;
}

.user-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: var(--space-xs);
  min-width: 180px;
  background-color: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  opacity: 0;
  visibility: hidden;
  transform: translateY(-8px);
  transition: all var(--transition-fast);
  z-index: 100;
}

.user-menu:hover .user-dropdown,
.user-menu:focus-within .user-dropdown {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  font-size: 0.875rem;
  color: var(--text-secondary);
  background: none;
  border: none;
  text-decoration: none;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dropdown-item:first-child {
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}

.dropdown-item:last-child {
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}

.dropdown-item:hover {
  color: var(--text-primary);
  background-color: var(--bg-card-hover);
}
</style>
