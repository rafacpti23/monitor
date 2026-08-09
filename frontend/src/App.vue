<template>
  <!-- Auth pages (login/register): render bare, no shell -->
  <RouterView v-if="isAuthPage" />

  <!-- App shell for everything else -->
  <div v-else class="app-layout">
    <Sidebar />
    <div class="main-content">
      <Topbar />
      <div class="page-container">
        <RouterView />
      </div>
      <AppFooter />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './components/Sidebar.vue'
import Topbar from './components/Topbar.vue'
import AppFooter from './components/AppFooter.vue'
import { RouterView } from 'vue-router'

const route = useRoute()
const isAuthPage = computed(() => route.name === 'Login' || route.name === 'Register')
</script>
