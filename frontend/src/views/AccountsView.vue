<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch } from '@/api'
import { formatAmount } from '@/format'
import CurrencyPicker from '@/components/CurrencyPicker.vue'

type Account = {
  id: number
  currency: string
  comment: string
  created_at: string
  balance: number
}

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
const deletingAccountId = ref<number | null>(null)

function loadAccounts() {
  apiFetch('/api/accounts')
    .then((response) => response.json())
    .then((data) => (accounts.value = data))
}

function loadRates() {
  apiFetch('/api/rates')
    .then((response) => response.json())
    .then((data) => (rates.value = data))
}

const availableCurrencies = computed(() => {
  return rates.value.map((r) => r.currency).sort()
})

function getRateForCurrency(currencyCode: string): number | null {
  const found = rates.value.find((r) => r.currency === currencyCode)
  return found ? found.rate_to_usd : null
}

const currencyTotals = computed(() => {
  const totals: Record<string, number> = {}
  accounts.value.forEach((a) => {
    totals[a.currency] = (totals[a.currency] ?? 0) + a.balance
  })
  return totals
})

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

const totalInAllCurrencies = computed(() => {
  const usdTotal = totalUSD.value
  const result: { currency: string; amount: number }[] = []
  for (const currency of Object.keys(currencyTotals.value)) {
    const rate = getRateForCurrency(currency)
    if (rate) {
      result.push({ currency, amount: usdTotal / rate })
    }
  }
  return result
})

let rateInterval: ReturnType<typeof setInterval>

onMounted(() => {
  loadAccounts()
  loadRates()
  rateInterval = setInterval(loadRates, 60000)
})

onUnmounted(() => {
  clearInterval(rateInterval)
})

async function addAccount() {
  const code = newCurrency.value.trim().toUpperCase()
  if (!code) return

  const response = await apiFetch('/api/accounts', {
    method: 'POST',
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

function confirmDeleteAccount(id: number) {
  deletingAccountId.value = id
}

function cancelDeleteAccount() {
  deletingAccountId.value = null
}

async function deleteAccount() {
  if (deletingAccountId.value === null) return
  const response = await apiFetch(`/api/accounts/${deletingAccountId.value}`, {
    method: 'DELETE',
  })
  deletingAccountId.value = null
  if (response.ok) {
    loadAccounts()
  }
}

function goToAccount(account: Account) {
  router.push(`/accounts/${account.id}`)
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Мои счета</h1>
    </div>

    <div class="content-grid">
      <!-- Основная колонка -->
      <div class="main-column">
        <!-- Форма создания счёта -->
        <div class="card">
          <h2 class="card-title">Новый счёт</h2>
          <form @submit.prevent="addAccount" class="add-form">
            <CurrencyPicker v-model="newCurrency" :currencies="availableCurrencies" />
            <input
              v-model="newComment"
              placeholder="Комментарий"
              class="input-field input-grow"
            />
            <button type="submit" class="btn btn-primary">Создать</button>
          </form>
        </div>

        <!-- Список счетов -->
        <div class="card">
          <div v-if="accounts.length === 0" class="empty-state">
            <p class="empty-icon">&#128176;</p>
            <p>Счетов пока нет. Создайте первый!</p>
          </div>

          <div class="accounts-list">
            <div
              v-for="account in accounts"
              :key="account.id"
              class="account-item"
              @click="goToAccount(account)"
            >
              <div class="account-left">
                <span class="account-currency">{{ account.currency }}</span>
                <span class="account-comment">{{ account.comment || '---' }}</span>
              </div>
              <div class="account-right">
                <span class="account-balance" :class="{ negative: account.balance < 0 }">
                  {{ formatAmount(account.balance) }}
                </span>
                <button
                  @click.stop="confirmDeleteAccount(account.id)"
                  class="btn-icon btn-danger"
                  title="Удалить"
                >
                  &times;
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Боковая колонка — итоги -->
      <div class="side-column" v-if="Object.keys(currencyTotals).length > 0">
        <div class="card summary-card">
          <h2 class="card-title">Итого</h2>
          <div class="summary-list">
            <div
              v-for="(amount, currency) in currencyTotals"
              :key="currency"
              class="summary-item"
            >
              <span class="summary-currency">{{ currency }}</span>
              <span class="summary-value" :class="{ negative: amount < 0 }">
                {{ formatAmount(amount) }}
              </span>
            </div>
          </div>

          <div
            v-for="item in totalInAllCurrencies"
            :key="item.currency"
            class="total-usd"
          >
            <span>Всего &approx; {{ item.currency }}</span>
            <span class="total-usd-value">{{ formatAmount(item.amount) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Модалка подтверждения удаления -->
    <Teleport to="body">
      <div v-if="deletingAccountId !== null" class="modal-overlay" @click.self="cancelDeleteAccount">
        <div class="modal-card">
          <p class="modal-text">Удалить счёт?</p>
          <div class="modal-buttons">
            <button class="btn btn-outline" @click="cancelDeleteAccount">Отмена</button>
            <button class="btn btn-danger-solid" @click="deleteAccount">Удалить</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.page {
  padding: 2rem;
  max-width: 1060px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 1.5rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #1a1a2e;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 1.5rem;
  align-items: start;
}

@media (max-width: 768px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}

.main-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.card {
  background: #fff;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.card-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 1rem;
}

.add-form {
  display: flex;
  gap: 0.75rem;
}

.input-field {
  padding: 0.65rem 0.9rem;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: border-color 0.2s;
  background: #fff;
  color: #333;
  width: 130px;
}

.input-grow {
  flex: 1;
  width: auto;
}

.input-field:focus {
  outline: none;
  border-color: #0f3460;
}

.btn {
  padding: 0.65rem 1.2rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: #0f3460;
  color: #fff;
}

.btn-primary:hover {
  background: #1a4a7a;
}

.empty-state {
  text-align: center;
  padding: 2rem 1rem;
  color: #999;
}

.empty-icon {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.accounts-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.9rem 1rem;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #eee;
  cursor: pointer;
  transition: all 0.15s;
}

.account-item:hover {
  background: #f0f4ff;
  border-color: #0f3460;
}

.account-left {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.account-currency {
  font-weight: 600;
  font-size: 0.95rem;
  color: #0f3460;
}

.account-comment {
  font-size: 0.8rem;
  color: #999;
}

.account-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.account-balance {
  font-size: 1.05rem;
  font-weight: 600;
  color: #22863a;
}

.account-balance.negative {
  color: #d73a49;
}

.btn-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  font-size: 1.1rem;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-danger {
  background: transparent;
  color: #ccc;
}

.btn-danger:hover {
  background: #fee;
  color: #d73a49;
}

.summary-card {
  position: sticky;
  top: 76px;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 0.8rem;
  background: #f0f4ff;
  border-radius: 8px;
  gap: 0.5rem;
}

.summary-currency {
  font-weight: 600;
  color: #0f3460;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.summary-value {
  font-weight: 700;
  font-size: 0.95rem;
  color: #22863a;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.summary-value.negative {
  color: #d73a49;
}

.total-usd {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  padding: 0.75rem 0.8rem;
  background: linear-gradient(135deg, #0f3460 0%, #16213e 100%);
  border-radius: 8px;
  color: #fff;
  font-weight: 600;
  font-size: 0.85rem;
  gap: 0.5rem;
}

.total-usd:first-of-type {
  margin-top: 0.75rem;
}

.total-usd-value {
  font-size: 0.95rem;
  font-weight: 700;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-card {
  background: #fff;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  min-width: 280px;
  text-align: center;
}

.modal-text {
  font-size: 1.05rem;
  font-weight: 500;
  color: #333;
  margin-bottom: 1.25rem;
}

.modal-buttons {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

.btn-outline {
  background: #fff;
  color: #333;
  border: 1.5px solid #ddd;
}

.btn-outline:hover {
  border-color: #bbb;
  background: #f9f9f9;
}

.btn-danger-solid {
  background: #d73a49;
  color: #fff;
}

.btn-danger-solid:hover {
  background: #b42d3a;
}
</style>
