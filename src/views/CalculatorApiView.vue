<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import LightLayout from '@/layouts/LightLayout.vue'

type Balance = {
  id?: number
  currency: string
  amount?: number
}

type Rate = {
  id: number
  currency: string
  rate_to_usd: number
  updated_at: string
}

const currencies = ref<Balance[]>([])
const rates = ref<Rate[]>([])
const newCurrency = ref('')
const newAmount = ref('')

function getRateForCurrency(currencyCode: string): number | null {
  const found = rates.value.find((r) => r.currency == currencyCode)
  return found ? found.rate_to_usd : null
}

const hasMissingRates = computed((): boolean => {
  return currencies.value.some((c) => getRateForCurrency(c.currency) == null)
})

const totalUSD = computed((): number => {
  let total = 0
  currencies.value.forEach((c) => {
    const rate = getRateForCurrency(c.currency)
    if (c.amount && rate) {
      total += c.amount * rate
    }
  })
  return total
})

const canCalculateTotal = computed((): boolean => {
  return !currencies.value.some((c) => {
    return c.amount && c.amount > 0 && getRateForCurrency(c.currency) === null
  })
})

function loadBalances() {
  fetch('/api/balances')
    .then((response) => response.json())
    .then((data) => (currencies.value = data))
}

function loadRates() {
  fetch('/api/rates')
    .then((response) => response.json())
    .then((data) => (rates.value = data))
}

onMounted(() => {
  loadBalances()
  loadRates()
  setInterval(loadRates, 5000)
})

async function addCurrency() {
  const code = newCurrency.value.trim().toUpperCase()
  if (!code) return

  const response = await fetch(`/api/balances`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      currency: code,
      amount: parseFloat(newAmount.value) || 0,
    }),
  })
  if (response.ok) {
    newCurrency.value = ''
    newAmount.value = ''
    loadBalances()
  }
}

async function removeBalance(balance: Balance) {
  const response = await fetch(`/api/balances/${balance.id}`, {
    method: 'DELETE',
  })

  if (response.ok) {
    loadBalances()
  }
}

async function updateBalance(balance: Balance) {
  await fetch(`/api/balances/${balance.id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(balance),
  })
}
</script>

<template>
  <LightLayout>
    <div class="calculator-container">
      <div class="calculator-card">
        <h1 class="calculator-title">Currency Calculator</h1>
        
        <div v-if="rates.length === 0" class="warning-banner">
          Курсы валют не загружены. Расчет невозможен.
        </div>
        <di v-else-if="hasMissingRates" class="warning-banner">
          Для некоторых валю нет курса. Общий расчет может быть неточным.
        </di>

        <form @submit.prevent="addCurrency" class="add-form">
          <input
            v-model="newCurrency"
            placeholder="Currency code (USD, EUR...)"
            class="input-field input-small"
          />
          <input v-model="newAmount" placeholder="Amount" class="input-field input-small" />

          <button type="submit" class="btn btn-primary">+ Add</button>
        </form>

        <div v-if="currencies.length === 0" class="empty-state">
          No currencies yet. Add one above!
        </div>

        <div class="currencies-list">
          <div v-for="balance in currencies" :key="balance.currency" class="currency-item">
            <span class="currency-code">{{ balance.currency }}</span>

            <input
              v-model="balance.amount"
              @change="updateBalance(balance)"
              placeholder="Amount"
              type="number"
              class="input-field input-small"
            />

            <span v-if="getRateForCurrency(balance.currency) !== null" class="rate-display">
              1 {{ balance.currency }} = {{ getRateForCurrency(balance.currency)!.toFixed(4) }} USD
            </span>

            <span v-else class="rate-missing"> нет курса </span>

            <button @click="removeBalance(balance)" class="btn btn-danger">✕</button>
          </div>
        </div>

        <div v-if="currencies.length > 0" class="total-section">
          <h2 class="total-title">Total Conversion</h2>

          <div v-if="canCalculateTotal" class="warning-banner warning-small">
            Не ве курсы доступны – итог может быть неполным
          </div>

          <ul class="total-list">
            <li
              v-if="!currencies.filter((c) => c.currency === 'USD').length && totalUSD"
              class="total-item"
            >
              <span class="total-currency">USD</span>
              <span class="total-value">{{ totalUSD.toFixed(2) }}</span>
            </li>
            <template v-for="currency in currencies" :key="currency.currency">
              <li v-if="getRateForCurrency(currency.currency)" class="total-item">
                <span class="total-currency">{{ currency.currency }}</span>
                <span class="total-value">
                  {{ (totalUSD / getRateForCurrency(currency.currency)!).toFixed(2) }}
                </span>
              </li>
            </template>
          </ul>
        </div>
      </div>
    </div>
  </LightLayout>
</template>

<style scoped>
/* Жёлтая плашка-предупреждение */
.warning-banner {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffc107;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  text-align: center;
}

.warning-small {
  margin-bottom: 0.75rem;
  font-size: 0.85rem;
}

/* Автоматический курс рядом с валютой */
.rate-display {
  font-size: 0.85rem;
  color: #28a745;
  font-weight: 500;
  white-space: nowrap;
}

/* Нет курса — красный текст */
.rate-missing {
  font-size: 0.85rem;
  color: #dc3545;
  font-weight: 500;
  font-style: italic;
}

.calculator-container {
  width: 100%;
  max-width: 500px;
}

.calculator-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  padding: 2rem;
}

.calculator-title {
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
  flex: 1;
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

.input-small {
  flex: none;
  width: 100px;
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

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #888;
  font-size: 1rem;
}

.currencies-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.currency-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 10px;
  border: 1px solid #e9ecef;
}

.currency-code {
  font-weight: 600;
  font-size: 1rem;
  color: #4a90d9;
  min-width: 50px;
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

.total-currency {
  font-weight: 600;
}

.total-value {
  font-size: 1.25rem;
  font-weight: 700;
}
</style>
