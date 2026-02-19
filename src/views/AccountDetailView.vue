<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LightLayout from '@/layouts/LightLayout.vue'

// Тип счёта — приходит с бэкенда, баланс вычисляется через SUM транзакций
type Account = {
  id: number
  currency: string
  comment: string
  created_at: string
  balance: number
}

// Тип транзакции (операции по счёту)
type Transaction = {
  id: number
  account_id: number
  amount: number
  comment: string
  created_at: string
}

const route = useRoute()
const router = useRouter()

const account = ref<Account | null>(null)
const transactions = ref<Transaction[]>([])
const txAmount = ref('')
const txComment = ref('')
const error = ref('')

// ID счёта из URL
const accountId = Number(route.params.id)

// Загрузка данных счёта
function loadAccount() {
  fetch(`/api/accounts/${accountId}`)
    .then((response) => {
      if (!response.ok) throw new Error('Счёт не найден')
      return response.json()
    })
    .then((data) => (account.value = data))
    .catch(() => (error.value = 'Счёт не найден'))
}

// Загрузка истории транзакций (новые сверху)
function loadTransactions() {
  fetch(`/api/accounts/${accountId}/transactions`)
    .then((response) => response.json())
    .then((data) => (transactions.value = data))
}

onMounted(() => {
  loadAccount()
  loadTransactions()
})

// Добавить операцию — пополнение (положительная сумма)
async function deposit() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return

  await createTransaction(amount)
}

// Добавить операцию — списание (отрицательная сумма)
async function withdraw() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return

  await createTransaction(-amount)
}

// Создание транзакции через API
async function createTransaction(amount: number) {
  const response = await fetch(`/api/accounts/${accountId}/transactions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      amount: amount,
      comment: txComment.value.trim(),
    }),
  })

  if (response.ok) {
    txAmount.value = ''
    txComment.value = ''
    loadAccount()
    loadTransactions()
  }
}

// Форматирование даты для отображения
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('ru-RU')
}

// Назад к списку счетов
function goBack() {
  router.push('/accounts')
}
</script>

<template>
  <LightLayout>
    <div class="detail-container">
      <div class="detail-card">
        <!-- Кнопка «Назад» -->
        <button @click="goBack" class="btn btn-back">← Назад к счетам</button>

        <!-- Ошибка -->
        <div v-if="error" class="error-banner">{{ error }}</div>

        <!-- Информация о счёте -->
        <div v-if="account" class="account-info-section">
          <h1 class="detail-title">
            {{ account.currency }}
            <span class="detail-comment" v-if="account.comment">— {{ account.comment }}</span>
          </h1>
          <div class="balance-display" :class="{ negative: account.balance < 0 }">
            Баланс: {{ account.balance.toFixed(2) }} {{ account.currency }}
          </div>

          <!-- Форма добавления операции -->
          <div class="transaction-form">
            <h2 class="form-title">Новая операция</h2>
            <div class="form-row">
              <input
                v-model="txAmount"
                placeholder="Сумма"
                type="number"
                step="0.01"
                min="0.01"
                class="input-field input-amount"
              />
              <input
                v-model="txComment"
                placeholder="Комментарий"
                class="input-field input-comment"
              />
            </div>
            <div class="form-buttons">
              <button @click="deposit" class="btn btn-deposit">Пополнить</button>
              <button @click="withdraw" class="btn btn-withdraw">Списать</button>
            </div>
          </div>

          <!-- История операций -->
          <div class="transactions-section">
            <h2 class="section-title">История операций</h2>

            <div v-if="transactions.length === 0" class="empty-state">
              Операций пока нет
            </div>

            <div class="transactions-list">
              <div
                v-for="tx in transactions"
                :key="tx.id"
                class="transaction-item"
                :class="{ income: tx.amount > 0, expense: tx.amount < 0 }"
              >
                <div class="tx-left">
                  <span class="tx-amount">
                    {{ tx.amount > 0 ? '+' : '' }}{{ tx.amount.toFixed(2) }}
                  </span>
                  <span class="tx-comment" v-if="tx.comment">{{ tx.comment }}</span>
                </div>
                <span class="tx-date">{{ formatDate(tx.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </LightLayout>
</template>

<style scoped>
.detail-container {
  width: 100%;
  max-width: 550px;
}

.detail-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  padding: 2rem;
}

.btn-back {
  background: none;
  border: none;
  color: #4a90d9;
  cursor: pointer;
  font-size: 0.95rem;
  padding: 0.5rem 0;
  margin-bottom: 1rem;
  font-weight: 500;
}

.btn-back:hover {
  color: #3a7bc8;
}

.error-banner {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  text-align: center;
}

.detail-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 0.5rem;
}

.detail-comment {
  font-weight: 400;
  color: #888;
  font-size: 1.25rem;
}

.balance-display {
  font-size: 1.5rem;
  font-weight: 700;
  color: #28a745;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 10px;
  text-align: center;
}

.balance-display.negative {
  color: #dc3545;
}

.transaction-form {
  margin-bottom: 2rem;
  padding: 1.25rem;
  background: #f0f4f8;
  border-radius: 10px;
}

.form-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 0.75rem;
}

.form-row {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
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

.input-amount {
  width: 120px;
  flex: none;
}

.input-comment {
  flex: 1;
}

.form-buttons {
  display: flex;
  gap: 0.75rem;
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

.btn-deposit {
  background: #28a745;
  color: #fff;
  flex: 1;
}

.btn-deposit:hover {
  background: #218838;
}

.btn-withdraw {
  background: #dc3545;
  color: #fff;
  flex: 1;
}

.btn-withdraw:hover {
  background: #c82333;
}

.transactions-section {
  border-top: 2px solid #e9ecef;
  padding-top: 1.5rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 1rem;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #888;
  font-size: 1rem;
}

.transactions-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.transaction-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  border-left: 4px solid;
}

.transaction-item.income {
  background: #f0fff4;
  border-color: #28a745;
}

.transaction-item.expense {
  background: #fff5f5;
  border-color: #dc3545;
}

.tx-left {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.tx-amount {
  font-weight: 600;
  font-size: 1rem;
}

.income .tx-amount {
  color: #28a745;
}

.expense .tx-amount {
  color: #dc3545;
}

.tx-comment {
  font-size: 0.85rem;
  color: #888;
}

.tx-date {
  font-size: 0.8rem;
  color: #aaa;
  white-space: nowrap;
}
</style>
