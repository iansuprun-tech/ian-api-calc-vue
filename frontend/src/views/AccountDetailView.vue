<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '@/api'

type Account = {
  id: number
  currency: string
  comment: string
  created_at: string
  balance: number
}

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

const accountId = Number(route.params.id)

function loadAccount() {
  apiFetch(`/api/accounts/${accountId}`)
    .then((response) => {
      if (!response.ok) throw new Error('Счёт не найден')
      return response.json()
    })
    .then((data) => (account.value = data))
    .catch(() => (error.value = 'Счёт не найден'))
}

function loadTransactions() {
  apiFetch(`/api/accounts/${accountId}/transactions`)
    .then((response) => response.json())
    .then((data) => (transactions.value = data))
}

onMounted(() => {
  loadAccount()
  loadTransactions()
})

async function deposit() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return
  await createTransaction(amount)
}

async function withdraw() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return
  await createTransaction(-amount)
}

async function createTransaction(amount: number) {
  const response = await apiFetch(`/api/accounts/${accountId}/transactions`, {
    method: 'POST',
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

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('ru-RU')
}

function goBack() {
  router.push('/accounts')
}
</script>

<template>
  <div class="page">
    <button @click="goBack" class="back-link">&larr; Назад к счетам</button>

    <div v-if="error" class="error-banner">{{ error }}</div>

    <div v-if="account" class="detail-grid">
      <!-- Карточка счёта -->
      <div class="card account-card">
        <div class="account-header">
          <div>
            <h1 class="account-name">{{ account.currency }}</h1>
            <p class="account-comment" v-if="account.comment">{{ account.comment }}</p>
          </div>
          <div class="balance-badge" :class="{ negative: account.balance < 0 }">
            {{ account.balance.toFixed(2) }} {{ account.currency }}
          </div>
        </div>
      </div>

      <!-- Форма операции -->
      <div class="card">
        <h2 class="card-title">Новая операция</h2>
        <div class="tx-form">
          <input
            v-model="txAmount"
            placeholder="Сумма"
            type="number"
            step="0.01"
            min="0.01"
            class="input-field"
          />
          <input
            v-model="txComment"
            placeholder="Комментарий"
            class="input-field input-grow"
          />
          <div class="tx-buttons">
            <button @click="deposit" class="btn btn-success">Пополнить</button>
            <button @click="withdraw" class="btn btn-danger">Списать</button>
          </div>
        </div>
      </div>

      <!-- История операций -->
      <div class="card">
        <h2 class="card-title">История операций</h2>

        <div v-if="transactions.length === 0" class="empty-state">
          Операций пока нет
        </div>

        <div class="tx-list">
          <div
            v-for="tx in transactions"
            :key="tx.id"
            class="tx-item"
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
</template>

<style scoped>
.page {
  padding: 2rem;
  max-width: 700px;
  margin: 0 auto;
}

.back-link {
  background: none;
  border: none;
  color: #0f3460;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 500;
  padding: 0;
  margin-bottom: 1.5rem;
  display: inline-block;
}

.back-link:hover {
  text-decoration: underline;
}

.error-banner {
  background: #fee;
  color: #c33;
  border: 1px solid #fcc;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  text-align: center;
}

.detail-grid {
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

.account-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.account-name {
  font-size: 1.75rem;
  font-weight: 700;
  color: #1a1a2e;
}

.account-comment {
  color: #888;
  font-size: 0.9rem;
  margin-top: 0.2rem;
}

.balance-badge {
  font-size: 1.3rem;
  font-weight: 700;
  color: #22863a;
  background: #f0fdf4;
  padding: 0.6rem 1.2rem;
  border-radius: 10px;
  white-space: nowrap;
}

.balance-badge.negative {
  color: #d73a49;
  background: #fef2f2;
}

.tx-form {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.input-field {
  padding: 0.65rem 0.9rem;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: border-color 0.2s;
  background: #fff;
  color: #333;
  width: 120px;
}

.input-grow {
  flex: 1;
  width: auto;
  min-width: 150px;
}

.input-field:focus {
  outline: none;
  border-color: #0f3460;
}

.tx-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn {
  padding: 0.65rem 1rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-success {
  background: #22863a;
  color: #fff;
}

.btn-success:hover {
  background: #1b6e30;
}

.btn-danger {
  background: #d73a49;
  color: #fff;
}

.btn-danger:hover {
  background: #b42d3a;
}

.empty-state {
  text-align: center;
  padding: 2rem 1rem;
  color: #999;
}

.tx-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.tx-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.7rem 0.9rem;
  border-radius: 8px;
  border-left: 3px solid;
}

.tx-item.income {
  background: #f0fdf4;
  border-color: #22863a;
}

.tx-item.expense {
  background: #fef2f2;
  border-color: #d73a49;
}

.tx-left {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.tx-amount {
  font-weight: 600;
  font-size: 0.95rem;
}

.income .tx-amount {
  color: #22863a;
}

.expense .tx-amount {
  color: #d73a49;
}

.tx-comment {
  font-size: 0.8rem;
  color: #888;
}

.tx-date {
  font-size: 0.75rem;
  color: #aaa;
  white-space: nowrap;
}
</style>
