<script setup lang="ts">
import { RouterLink } from "vue-router"
</script>

<template>
<div class="light-layout">
  <header class="light-header">
    <nav class="light-nav">
      <RouterLink to="/" class="nav-link">Home</RouterLink>
      <RouterLink to="/about" class="nav-link">About</RouterLink>
      <RouterLink to="/accounts" class="nav-link">Счета</RouterLink>
    </nav>
  </header>

  <main class="light-main">
    <slot />
  </main>

</div>
</template>

<style scoped>
.light-layout {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
  color: #333;
}

.light-header {
  background: #fff;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1rem 2rem;
}

.light-nav {
  display: flex;
  gap: 2rem;
  justify-content: center;
  align-items: center;
}

.nav-link {
  color: #555;
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.nav-link:hover {
  background: #f0f0f0;
  color: #333;
}

.nav-link.router-link-exact-active {
  background: #4a90d9;
  color: #fff;
}

.light-main {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 3rem 1rem;
  min-height: calc(100vh - 70px);
}
</style>
