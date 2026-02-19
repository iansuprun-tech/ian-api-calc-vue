<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import LightLayout from '@/layouts/LightLayout.vue'

// Тип счёта — приходит с бэкенда, баланс вычисляется через SUM транзакций
type Account = {
  id: number
  currency: string
  comment: string
  created_at: string
  balance: number
}

// Тип курса валют (для расчёта итогов в USD)
type Rate = {
  id: number
  currency: string
  rate_to_usd: number
  updated_at: string
}

const router = useRouter()

const accounts = ref<Account[]>([])
const rates = ref<Rate[]>([])
const newCurrency = ref('')
const newComment = ref('')

// Загрузка списка счетов с бэкенда
function loadAccounts() {
  fetch('/api/accounts')
    .then((response) => response.json())
    .then((data) => (accounts.value = data))
}

// Загрузка курсов валют
function loadRates() {
  fetch('/api/rates')
    .then((response) => response.json())
    .then((data) => (rates.value = data))
}

// Получить курс валюты к USD
function getRateForCurrency(currencyCode: string): number | null {
  const found = rates.value.find((r) => r.currency === currencyCode)
  return found ? found.rate_to_usd : null
}

// Итоги по валютам — суммируем балансы всех счетов в одной валюте
const currencyTotals = computed(() => {
  const totals: Record<string, number> = {}
  accounts.value.forEach((a) => {
    totals[a.currency] = (totals[a.currency] ?? 0) + a.balance
  })
  return totals
})

// Общий итог в USD (для расчёта через курсы)
const totalUSD = computed((): number => {
  let total = 0
  for (const [currency, amount] of Object.entries(currencyTotals.value)) {
    const rate = getRateForCurrency(currency)
    if (rate) {
      total += amount * rate
    }
  }
  return total
})

onMounted(() => {
  loadAccounts()
  loadRates()
  setInterval(loadRates, 5000)
})

// Создание нового счёта
async function addAccount() {
  const code = newCurrency.value.trim().toUpperCase()
  if (!code) return

  const response = await fetch('/api/accounts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      currency: code,
      comment: newComment.value.trim(),
    }),
  })
  if (response.ok) {
    newCurrency.value = ''
    newComment.value = ''
    loadAccounts()
  }
}

// Удаление счёта
async function removeAccount(account: Account) {
  const response = await fetch(`/api/accounts/${account.id}`, {
    method: 'DELETE',
  })
  if (response.ok) {
    loadAccounts()
  }
}

// Переход на страницу счёта
function goToAccount(account: Account) {
  router.push(`/accounts/${account.id}`)
}
</script>

<template>
  <LightLayout>
    <div class="accounts-container">
      <div class="accounts-card">
        <h1 class="accounts-title">Счета</h1>

        <!-- Форма создания нового счёта -->
        <form @submit.prevent="addAccount" class="add-form">
          <input
            v-model="newCurrency"
            placeholder="Валюта (USD, EUR...)"
            class="input-field input-currency"
          />
          <input
            v-model="newComment"
            placeholder="Комментарий"
            class="input-field input-comment"
          />
          <button type="submit" class="btn btn-primary">+ Создать</button>
        </form>

        <!-- Пустое состояние -->
        <div v-if="accounts.length === 0" class="empty-state">
          Счетов пока нет. Создайте первый!
        </div>

        <!-- Список счетов -->
        <div class="accounts-list">
          <div
            v-for="account in accounts"
            :key="account.id"
            class="account-item"
            @click="goToAccount(account)"
          >
            <div class="account-info">
              <span class="account-currency">{{ account.currency }}</span>
              <span class="account-comment">{{ account.comment || '—' }}</span>
            </div>
            <div class="account-right">
              <span class="account-balance" :class="{ negative: account.balance < 0 }">
                {{ account.balance.toFixed(2) }}
              </span>
              <button
                @click.stop="removeAccount(account)"
                class="btn btn-danger btn-small"
              >
                ✕
              </button>
            </div>
          </div>
        </div>

        <!-- Итоги по валютам -->
        <div v-if="Object.keys(currencyTotals).length > 0" class="total-section">
          <h2 class="total-title">Итого по валютам</h2>
          <ul class="total-list">
            <li
              v-for="(amount, currency) in currencyTotals"
              :key="currency"
              class="total-item"
            >
              <span class="total-currency">{{ currency }}</span>
              <span class="total-value">{{ amount.toFixed(2) }}</span>
            </li>
          </ul>

          <!-- Итого в USD (если есть курсы) -->
          <div v-if="totalUSD !== 0" class="total-usd">
            <span class="total-currency">Всего ≈ USD</span>
            <span class="total-value">{{ totalUSD.toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </div>
  </LightLayout>
</template>

<style scoped>
.accounts-container {
  width: 100%;
  max-width: 550px;
}

.accounts-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  padding: 2rem;
}

.accounts-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: #333;
  text-align: center;
  margin-bottom: 1.5rem;
}

.add-form {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.input-field {
  padding: 0.75rem 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
  background: #fff;
  color: #333;
}

.input-field:focus {
  outline: none;
  border-color: #4a90d9;
}

.input-currency {
  width: 110px;
  flex: none;
}

.input-comment {
  flex: 1;
}

.btn {
  padding: 0.75rem 1.25rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #4a90d9;
  color: #fff;
  white-space: nowrap;
}

.btn-primary:hover {
  background: #3a7bc8;
}

.btn-danger {
  background: #ff6b6b;
  color: #fff;
  padding: 0.5rem 0.75rem;
}

.btn-danger:hover {
  background: #ee5a5a;
}

.btn-small {
  padding: 0.3rem 0.6rem;
  font-size: 0.85rem;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #888;
  font-size: 1rem;
}

.accounts-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 10px;
  border: 1px solid #e9ecef;
  cursor: pointer;
  transition: all 0.2s;
}

.account-item:hover {
  background: #eef2f7;
  border-color: #4a90d9;
}

.account-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.account-currency {
  font-weight: 600;
  font-size: 1rem;
  color: #4a90d9;
}

.account-comment {
  font-size: 0.85rem;
  color: #888;
}

.account-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.account-balance {
  font-size: 1.1rem;
  font-weight: 600;
  color: #28a745;
}

.account-balance.negative {
  color: #dc3545;
}

.total-section {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid #e9ecef;
}

.total-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 1rem;
}

.total-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.total-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  margin-bottom: 0.5rem;
  color: #fff;
}

.total-usd {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border-radius: 8px;
  margin-top: 0.5rem;
  color: #fff;
  font-weight: 600;
}

.total-currency {
  font-weight: 600;
}

.total-value {
  font-size: 1.25rem;
  font-weight: 700;
}
</style>
