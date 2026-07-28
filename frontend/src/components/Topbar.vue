<template>
  <header class="topbar">
    <div class="topbar-title">{{ pageTitle }}</div>
    
    <div class="topbar-status">
      <div class="status-dot" :class="{ offline: !connected }"></div>
      <span>{{ connected ? 'Conectado' : 'Reconectando...' }}</span>
    </div>

    <div style="margin-left: auto; display: flex; align-items: center; gap: 12px;">
      <!-- Notifications -->
      <button class="btn btn-ghost btn-sm" style="position: relative;">
        <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M10 2a6 6 0 0 1 6 6v3l1.5 2.5H2.5L4 11V8a6 6 0 0 1 6-6z"/>
          <path d="M8 16a2 2 0 0 0 4 0"/>
        </svg>
        <span v-if="notifCount > 0" style="position: absolute; top: 2px; right: 2px; width: 8px; height: 8px; background: var(--accent-red); border-radius: 50%;"></span>
      </button>

      <!-- User dropdown -->
      <div class="user-dropdown-wrapper" ref="dropdownRef">
        <button class="topbar-user" @click="showDropdown = !showDropdown">
          <img v-if="gravatarUrl" :src="gravatarUrl" alt="avatar" class="user-avatar-img" />
          <div v-else class="user-avatar">{{ userInitials }}</div>
          <div>
            <div class="user-name">{{ userName }}</div>
            <div class="user-role">{{ userRole }}</div>
          </div>
          <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" style="color: var(--text-muted)">
            <path d="M6 8l4 4 4-4"/>
          </svg>
        </button>

        <Transition name="dropdown">
          <div v-if="showDropdown" class="user-dropdown">
            <div class="dropdown-header">
              <img v-if="gravatarUrl" :src="gravatarUrl" alt="avatar" class="dropdown-avatar-img" />
              <div v-else class="dropdown-avatar">{{ userInitials }}</div>
              <div>
                <div class="dropdown-name">{{ userName }}</div>
                <div class="dropdown-email">{{ userEmail }}</div>
              </div>
            </div>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item" @click="goSettings">
              <svg width="15" height="15" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="10" cy="10" r="2"/>
                <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.22 4.22l1.42 1.42M14.36 14.36l1.42 1.42M4.22 15.78l1.42-1.42M14.36 5.64l1.42-1.42"/>
              </svg>
              Configurações do Perfil
            </button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item dropdown-item-danger" @click="handleLogout">
              <svg width="15" height="15" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M7 4H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3M10 10h7M14 7l3 3-3 3"/>
              </svg>
              Sair
            </button>
          </div>
        </Transition>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const connected = ref(true)
const notifCount = ref(0)
const showDropdown = ref(false)
const dropdownRef = ref(null)

const userName = computed(() => auth.user?.name || '—')
const userRole = computed(() => auth.user?.role || '')
const userEmail = computed(() => auth.user?.email || '')

const userInitials = computed(() => {
  const name = auth.user?.name || ''
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) || '·'
})

// Gravatar URL from email MD5 hash
const gravatarUrl = computed(() => {
  const email = auth.user?.email
  if (!email) return ''
  return `https://www.gravatar.com/avatar/${md5(email.trim().toLowerCase())}?s=80&d=404`
})

// Lightweight MD5 for Gravatar (only needs email hashing)
function md5(str) {
  function rotl(v, s) { return (v << s) | (v >>> (32 - s)) }
  function cmn(q, a, b, x, s, t) { a = (a + q + (x >>> 0) + t) | 0; return ((rotl(a, s) + b) | 0) }
  function ff(a, b, c, d, x, s, t) { return cmn((b & c) | (~b & d), a, b, x, s, t) }
  function gg(a, b, c, d, x, s, t) { return cmn((b & d) | (c & ~d), a, b, x, s, t) }
  function hh(a, b, c, d, x, s, t) { return cmn(b ^ c ^ d, a, b, x, s, t) }
  function ii(a, b, c, d, x, s, t) { return cmn(c ^ (b | ~d), a, b, x, s, t) }
  const bytes = []
  for (let i = 0; i < str.length; i++) {
    const c = str.charCodeAt(i)
    if (c < 128) bytes.push(c)
    else if (c < 2048) { bytes.push(192 | (c >> 6)); bytes.push(128 | (c & 63)) }
    else { bytes.push(224 | (c >> 12)); bytes.push(128 | ((c >> 6) & 63)); bytes.push(128 | (c & 63)) }
  }
  const n = bytes.length
  bytes.push(0x80)
  while (bytes.length % 64 !== 56) bytes.push(0)
  const bits = n * 8
  bytes.push(bits & 0xff, (bits >> 8) & 0xff, (bits >> 16) & 0xff, (bits >> 24) & 0xff, 0, 0, 0, 0)
  let a = 0x67452301, b = 0xEFCDAB89, c = 0x98BADCFE, d = 0x10325476
  for (let i = 0; i < bytes.length; i += 64) {
    const w = []
    for (let j = 0; j < 16; j++) w[j] = bytes[i+j*4] | (bytes[i+j*4+1]<<8) | (bytes[i+j*4+2]<<16) | (bytes[i+j*4+3]<<24)
    let aa=a, bb=b, cc=c, dd=d
    a=ff(a,b,c,d,w[0],7,-680876936);d=ff(d,a,b,c,w[1],12,-389564586);c=ff(c,d,a,b,w[2],17,606105819);b=ff(b,c,d,a,w[3],22,-1044525330)
    a=ff(a,b,c,d,w[4],7,-176418897);d=ff(d,a,b,c,w[5],12,1200080426);c=ff(c,d,a,b,w[6],17,-1473231341);b=ff(b,c,d,a,w[7],22,-45705983)
    a=ff(a,b,c,d,w[8],7,1770035416);d=ff(d,a,b,c,w[9],12,-1958414417);c=ff(c,d,a,b,w[10],17,-42063);b=ff(b,c,d,a,w[11],22,-1990404162)
    a=ff(a,b,c,d,w[12],7,1804603682);d=ff(d,a,b,c,w[13],12,-40341101);c=ff(c,d,a,b,w[14],17,-1502002290);b=ff(b,c,d,a,w[15],22,1236535329)
    a=gg(a,b,c,d,w[1],5,-165796510);d=gg(d,a,b,c,w[6],9,-1069501632);c=gg(c,d,a,b,w[11],14,643717713);b=gg(b,c,d,a,w[0],20,-373897302)
    a=gg(a,b,c,d,w[5],5,-701558691);d=gg(d,a,b,c,w[10],9,38016083);c=gg(c,d,a,b,w[15],14,-660478335);b=gg(b,c,d,a,w[4],20,-405537848)
    a=gg(a,b,c,d,w[9],5,568446438);d=gg(d,a,b,c,w[14],9,-1019803690);c=gg(c,d,a,b,w[3],14,-187363961);b=gg(b,c,d,a,w[8],20,1163531501)
    a=gg(a,b,c,d,w[13],5,-1444681467);d=gg(d,a,b,c,w[2],9,-51403784);c=gg(c,d,a,b,w[7],14,1735328473);b=gg(b,c,d,a,w[12],20,-1926607734)
    a=hh(a,b,c,d,w[5],4,-378558);d=hh(d,a,b,c,w[8],11,-2022574463);c=hh(c,d,a,b,w[11],16,1839030562);b=hh(b,c,d,a,w[14],23,-35309556)
    a=hh(a,b,c,d,w[1],4,-1530992060);d=hh(d,a,b,c,w[4],11,1272893353);c=hh(c,d,a,b,w[7],16,-155497632);b=hh(b,c,d,a,w[10],23,-1094730640)
    a=hh(a,b,c,d,w[13],4,681279174);d=hh(d,a,b,c,w[0],11,-358537222);c=hh(c,d,a,b,w[3],16,-722521979);b=hh(b,c,d,a,w[6],23,76029189)
    a=hh(a,b,c,d,w[9],4,-640364487);d=hh(d,a,b,c,w[12],11,-421815835);c=hh(c,d,a,b,w[15],16,530742520);b=hh(b,c,d,a,w[2],23,-995338651)
    a=ii(a,b,c,d,w[0],6,-198630844);d=ii(d,a,b,c,w[7],10,1126891415);c=ii(c,d,a,b,w[14],15,-1416354905);b=ii(b,c,d,a,w[5],21,-57434055)
    a=ii(a,b,c,d,w[12],6,1700485571);d=ii(d,a,b,c,w[3],10,-1894986606);c=ii(c,d,a,b,w[10],15,-1051523);b=ii(b,c,d,a,w[1],21,-2054922799)
    a=ii(a,b,c,d,w[8],6,1873313359);d=ii(d,a,b,c,w[15],10,-30611744);c=ii(c,d,a,b,w[6],15,-1560198380);b=ii(b,c,d,a,w[13],21,1309151649)
    a=ii(a,b,c,d,w[4],6,-145523070);d=ii(d,a,b,c,w[11],10,-1120210379);c=ii(c,d,a,b,w[2],15,718787259);b=ii(b,c,d,a,w[9],21,-343485551)
    a=(a+aa)|0;b=(b+bb)|0;c=(c+cc)|0;d=(d+dd)|0
  }
  const hex = x => {
    let s = ''
    for (let i = 0; i < 4; i++) s += ('0' + ((x >> (i * 8)) & 0xff).toString(16)).slice(-2)
    return s
  }
  return hex(a) + hex(b) + hex(c) + hex(d)
}

const gravatarError = ref(false)

const pageTitle = computed(() => {
  const titles = {
    '/dashboard': 'Dashboard',
    '/servers': 'Servidores',
    '/websites': 'Websites',
    '/checks': 'Checks',
    '/incidents': 'Incidentes',
    '/settings': 'Configurações',
    '/company': 'Empresa',
    '/api-docs': 'Documentação da API',
  }
  return titles[route.path] || (auth.company?.system_name || 'P-mon')
})

function goSettings() {
  showDropdown.value = false
  router.push('/settings')
}

function handleLogout() {
  showDropdown.value = false
  auth.logout()
  router.push('/login')
}

// Close dropdown on click outside
function onClickOutside(e) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    showDropdown.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.user-dropdown-wrapper {
  position: relative;
}
.topbar-user {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-md, 8px);
  padding: 4px 8px;
  transition: all 0.15s;
  color: inherit;
  font-family: inherit;
}
.topbar-user:hover {
  background: var(--bg-hover, rgba(255,255,255,0.04));
  border-color: var(--border);
}
.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--accent-glow, rgba(0,230,118,0.12));
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 12px;
  flex-shrink: 0;
}
.user-avatar-img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}
.user-name { font-size: 13px; font-weight: 500; text-align: left; }
.user-role { font-size: 11px; color: var(--text-muted); text-align: left; }

/* Dropdown */
.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 260px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border-bright, #30363d);
  border-radius: var(--radius-lg, 12px);
  box-shadow: 0 16px 48px rgba(0,0,0,0.5);
  z-index: 1000;
  overflow: hidden;
}
.dropdown-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
}
.dropdown-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--accent-glow, rgba(0,230,118,0.12));
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}
.dropdown-avatar-img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}
.dropdown-name { font-size: 14px; font-weight: 600; }
.dropdown-email { font-size: 12px; color: var(--text-muted); }
.dropdown-divider {
  height: 1px;
  background: var(--border, #21262d);
  margin: 0;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  background: none;
  border: none;
  color: var(--text-secondary, #c9d1d9);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.12s;
  text-align: left;
  font-family: inherit;
}
.dropdown-item:hover {
  background: var(--bg-hover, rgba(255,255,255,0.04));
  color: var(--text-primary, #f0f6fc);
}
.dropdown-item-danger:hover {
  background: rgba(244,67,54,0.1);
  color: var(--accent-red, #f44336);
}

/* Transition */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
