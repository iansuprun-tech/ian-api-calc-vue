<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="app" :class="{ 'with-nav': auth.isAuthenticated }">
    <header v-if="auth.isAuthenticated" class="topbar">
      <div class="topbar-left">
        <h1 class="topbar-logo">FinTrack</h1>
        <nav class="topbar-nav">
          <RouterLink to="/accounts" class="nav-link">Счета</RouterLink>
          <RouterLink to="/categories" class="nav-link">Категории</RouterLink>
        </nav>
      </div>
      <button @click="logout" class="logout-btn">Выйти</button>
    </header>

    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
}

.app.with-nav {
  display: flex;
  flex-direction: column;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  height: 60px;
  background: #0f3460;
  color: #fff;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
  flex-shrink: 0;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.topbar-logo {
  font-size: 1.3rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.topbar-nav {
  display: flex;
  gap: 0.5rem;
}

.nav-link {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-weight: 500;
  font-size: 0.95rem;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.nav-link:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.nav-link.router-link-active {
  color: #fff;
  background: rgba(255, 255, 255, 0.15);
}

.logout-btn {
  padding: 0.4rem 1rem;
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.main-content {
  flex: 1;
}
</style>
