import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { useAuthStore } from './stores/auth'
import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// Hydrate user/company from token before mounting so the shell renders
// with the correct name/role/branding instead of "Guest / viewer".
const auth = useAuthStore(pinia)
auth.fetchMe().finally(() => {
  app.mount('#app')
})
