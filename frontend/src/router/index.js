import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/servers',
    name: 'Servers',
    component: () => import('../views/Servers.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/servers/:id',
    name: 'ServerDetail',
    component: () => import('../views/Servers/Detail.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/add-server',
    name: 'AddServer',
    component: () => import('../views/AddServer.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/websites',
    name: 'Websites',
    component: () => import('../views/Websites.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/websites/:id',
    name: 'WebsitesDetail',
    component: () => import('../views/Websites/Detail.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/checks',
    name: 'Checks',
    component: () => import('../views/Checks.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/incidents',
    name: 'Incidents',
    component: () => import('../views/Incidents.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/company',
    name: 'Company',
    component: () => import('../views/Company.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/api-docs',
    name: 'ApiDocs',
    component: () => import('../views/ApiDocs.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth guard: requiresAuth routes redirect to /login when no token.
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('p-mon-token')
  if (to.meta.requiresAuth && !token) {
    return next({ name: 'Login', query: { redirect: to.fullPath } })
  }
  // Authenticated users hitting /login or /register go to dashboard.
  if ((to.name === 'Login' || to.name === 'Register') && token) {
    return next({ name: 'Dashboard' })
  }
  next()
})

export default router
