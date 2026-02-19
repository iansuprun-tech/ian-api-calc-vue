<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')

async function register() {
  error.value = ''

  if (password.value !== confirmPassword.value) {
    error.value = 'Пароли не совпадают'
    return
  }

  if (password.value.length < 6) {
    error.value = 'Пароль должен быть не менее 6 символов'
    return
  }

  const response = await fetch('/api/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.value, password: password.value }),
  })

  if (!response.ok) {
    const data = await response.json()
    error.value = data.error || 'Ошибка регистрации'
    return
  }

  // Автоматический вход после регистрации
  const loginResponse = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.value, password: password.value }),
  })

  if (loginResponse.ok) {
    const data = await loginResponse.json()
    auth.setToken(data.token)
    router.push('/accounts')
  } else {
    router.push('/login')
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <h1 class="auth-logo">FinTrack</h1>
        <p class="auth-subtitle">Управление финансами</p>
      </div>

      <h2 class="auth-title">Регистрация</h2>

      <div v-if="error" class="auth-error">{{ error }}</div>

      <form @submit.prevent="register" class="auth-form">
        <div class="form-group">
          <label for="email">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            required
            autocomplete="email"
          />
        </div>

        <div class="form-group">
          <label for="password">Пароль</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Минимум 6 символов"
            required
            autocomplete="new-password"
          />
        </div>

        <div class="form-group">
          <label for="confirm-password">Подтверждение пароля</label>
          <input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            placeholder="Повторите пароль"
            required
            autocomplete="new-password"
          />
        </div>

        <button type="submit" class="auth-btn">Зарегистрироваться</button>
      </form>

      <p class="auth-link">
        Уже есть аккаунт?
        <RouterLink to="/login">Войти</RouterLink>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  padding: 1rem;
}

.auth-card {
  background: #fff;
  border-radius: 16px;
  padding: 2.5rem;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}

.auth-logo {
  font-size: 2rem;
  font-weight: 700;
  color: #0f3460;
  margin-bottom: 0.25rem;
}

.auth-subtitle {
  color: #888;
  font-size: 0.9rem;
}

.auth-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 1.5rem;
}

.auth-error {
  background: #fee;
  color: #c33;
  border: 1px solid #fcc;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 500;
  color: #555;
}

.form-group input {
  padding: 0.75rem 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
  background: #fff;
  color: #333;
}

.form-group input:focus {
  outline: none;
  border-color: #0f3460;
}

.auth-btn {
  padding: 0.85rem;
  background: #0f3460;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 0.5rem;
}

.auth-btn:hover {
  background: #1a4a7a;
}

.auth-link {
  text-align: center;
  margin-top: 1.5rem;
  color: #888;
  font-size: 0.9rem;
}

.auth-link a {
  color: #0f3460;
  font-weight: 500;
  text-decoration: none;
}

.auth-link a:hover {
  text-decoration: underline;
}
</style>
