]<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '@/api'
import { formatAmount } from '@/format'

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
  category_id: number | null
  category: string
  created_at: string
}

type Category = {
  id: number
  name: string
}

const route = useRoute()
const router = useRouter()

const account = ref<Account | null>(null)
const transactions = ref<Transaction[]>([])
const txAmount = ref('')
const txComment = ref('')
const txDate = ref('')
const selectedCategoryId = ref<number | null>(null)
const categories = ref<Category[]>([])
const error = ref('')
const deletingTxId = ref<number | null>(null)
const editingTx = ref<Transaction | null>(null)
const editingComment = ref(false)
const editCommentValue = ref('')

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

function loadCategories() {
  apiFetch('/api/categories')
    .then((response) => response.json())
    .then((data) => (categories.value = data))
}

function toggleCategory(id: number) {
  selectedCategoryId.value = selectedCategoryId.value === id ? null : id
}

onMounted(() => {
  loadAccount()
  loadTransactions()
  loadCategories()
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
  const body: Record<string, unknown> = {
    amount: amount,
    comment: txComment.value.trim(),
  }
  if (txDate.value) {
    body.created_at = new Date(txDate.value).toISOString()
  }
  if (selectedCategoryId.value !== null) {
    body.category_id = selectedCategoryId.value
  }
  const response = await apiFetch(`/api/accounts/${accountId}/transactions`, {
    method: 'POST',
    body: JSON.stringify(body),
  })

  if (response.ok) {
    txAmount.value = ''
    txComment.value = ''
    txDate.value = ''
    selectedCategoryId.value = null
    loadAccount()
    loadTransactions()
  }
}

function startEditComment() {
  if (!account.value) return
  editCommentValue.value = account.value.comment
  editingComment.value = true
}

function cancelEditComment() {
  editingComment.value = false
}

async function saveComment() {
  const response = await apiFetch(`/api/accounts/${accountId}`, {
    method: 'PUT',
    body: JSON.stringify({ comment: editCommentValue.value.trim() }),
  })
  if (response.ok) {
    const data = await response.json()
    account.value = data
    editingComment.value = false
  }
}

function startEdit(tx: Transaction) {
  editingTx.value = tx
  txAmount.value = Math.abs(tx.amount).toString()
  txComment.value = tx.comment
  selectedCategoryId.value = tx.category_id
  if (tx.created_at) {
    const d = new Date(tx.created_at.replace(' ', 'T'))
    if (!isNaN(d.getTime())) {
      const pad = (n: number) => n.toString().padStart(2, '0')
      txDate.value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
  }
}

function cancelEdit() {
  editingTx.value = null
  txAmount.value = ''
  txComment.value = ''
  txDate.value = ''
  selectedCategoryId.value = null
}

async function saveEdit() {
  if (!editingTx.value) return
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return

  const sign = editingTx.value.amount < 0 ? -1 : 1
  const body: Record<string, unknown> = {
    amount: amount * sign,
    comment: txComment.value.trim(),
  }
  if (txDate.value) {
    body.created_at = new Date(txDate.value).toISOString()
  } else {
    body.created_at = editingTx.value.created_at
  }
  if (selectedCategoryId.value !== null) {
    body.category_id = selectedCategoryId.value
  }

  const response = await apiFetch(
    `/api/accounts/${accountId}/transactions/${editingTx.value.id}`,
    {
      method: 'PUT',
      body: JSON.stringify(body),
    },
  )

  if (response.ok) {
    cancelEdit()
    loadAccount()
    loadTransactions()
  }
}

function confirmDelete(txId: number) {
  deletingTxId.value = txId
}

function cancelDelete() {
  deletingTxId.value = null
}

async function deleteTransaction() {
  if (deletingTxId.value === null) return
  const response = await apiFetch(`/api/accounts/${accountId}/transactions/${deletingTxId.value}`, {
    method: 'DELETE',
  })
  deletingTxId.value = null
  if (response.ok) {
    loadAccount()
    loadTransactions()
  }
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr.replace(' ', 'T'))
  if (isNaN(date.getTime())) return dateStr
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
            <div v-if="editingComment" class="comment-edit">
              <input
                v-model="editCommentValue"
                class="input-field input-grow"
                placeholder="Описание счёта"
                @keyup.enter="saveComment"
                @keyup.escape="cancelEditComment"
              />
              <button class="btn btn-success btn-sm" @click="saveComment">Сохранить</button>
              <button class="btn btn-outline btn-sm" @click="cancelEditComment">Отмена</button>
            </div>
            <div v-else class="comment-row">
              <p class="account-comment">{{ account.comment || '---' }}</p>
              <button class="btn-edit-comment" @click="startEditComment" title="Редактировать">&#9998;</button>
            </div>
          </div>
          <div class="balance-badge" :class="{ negative: account.balance < 0 }">
            {{ formatAmount(account.balance) }} {{ account.currency }}
          </div>
        </div>
      </div>

      <!-- Форма операции -->
      <div class="card">
        <h2 class="card-title">{{ editingTx ? 'Редактирование операции' : 'Новая операция' }}</h2>
        <div class="tx-form">
          <div v-if="categories.length > 0" class="category-buttons">
            <button
              v-for="cat in categories"
              :key="cat.id"
              class="category-btn"
              :class="{ active: selectedCategoryId === cat.id }"
              @click="toggleCategory(cat.id)"
            >
              {{ cat.name }}
            </button>
          </div>
          <div class="tx-row">
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
          </div>
          <div class="tx-row">
            <input
              v-model="txDate"
              type="datetime-local"
              class="input-field input-grow"
              title="Дата операции (необязательно)"
            />
            <div class="tx-buttons" v-if="editingTx">
              <button @click="saveEdit" class="btn btn-warning">Сохранить</button>
              <button @click="cancelEdit" class="btn btn-outline">Отмена</button>
            </div>
            <div class="tx-buttons" v-else>
              <button @click="deposit" class="btn btn-success">Пополнить</button>
              <button @click="withdraw" class="btn btn-danger">Списать</button>
            </div>
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
                {{ tx.amount > 0 ? '+' : '' }}{{ formatAmount(tx.amount) }}
              </span>
              <span class="tx-category" v-if="tx.category">{{ tx.category }}</span>
              <span class="tx-comment" v-if="tx.comment">{{ tx.comment }}</span>
            </div>
            <div class="tx-right">
              <span class="tx-date">{{ formatDate(tx.created_at) }}</span>
              <button class="btn-edit" @click="startEdit(tx)" title="Редактировать">&#9998;</button>
              <button class="btn-delete" @click="confirmDelete(tx.id)">&times;</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Модалка подтверждения удаления -->
    <Teleport to="body">
      <div v-if="deletingTxId !== null" class="modal-overlay" @click.self="cancelDelete">
        <div class="modal-card">
          <p class="modal-text">Удалить операцию?</p>
          <div class="modal-buttons">
            <button class="btn btn-outline" @click="cancelDelete">Отмена</button>
            <button class="btn btn-danger" @click="deleteTransaction">Удалить</button>
          </div>
        </div>
      </div>
    </Teleport>
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
  margin: 0;
}

.comment-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.2rem;
}

.comment-edit {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.3rem;
}

.btn-edit-comment {
  background: none;
  border: none;
  color: #aaa;
  font-size: 0.85rem;
  cursor: pointer;
  padding: 0;
  transition: color 0.2s;
}

.btn-edit-comment:hover {
  color: #d97706;
}

.btn-sm {
  padding: 0.35rem 0.7rem;
  font-size: 0.85rem;
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
  flex-direction: column;
  gap: 0.75rem;
}

.tx-row {
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

.tx-category {
  font-size: 0.75rem;
  font-weight: 500;
  color: #0f3460;
  background: #e8eef6;
  padding: 0.1rem 0.5rem;
  border-radius: 10px;
  width: fit-content;
}

.tx-comment {
  font-size: 0.8rem;
  color: #888;
}

.tx-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tx-date {
  font-size: 0.75rem;
  color: #aaa;
  white-space: nowrap;
}

.btn-warning {
  background: #d97706;
  color: #fff;
}

.btn-warning:hover {
  background: #b45309;
}

.btn-edit {
  background: none;
  border: 1px solid #ddd;
  border-radius: 6px;
  color: #999;
  font-size: 0.9rem;
  line-height: 1;
  width: 28px;
  height: 28px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-edit:hover {
  background: #d97706;
  border-color: #d97706;
  color: #fff;
}

.btn-delete {
  background: none;
  border: 1px solid #ddd;
  border-radius: 6px;
  color: #999;
  font-size: 1.1rem;
  line-height: 1;
  width: 28px;
  height: 28px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-delete:hover {
  background: #d73a49;
  border-color: #d73a49;
  color: #fff;
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

.category-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.category-btn {
  padding: 0.4rem 0.8rem;
  border: 1.5px solid #d0d7de;
  border-radius: 20px;
  background: #f0f4f8;
  color: #333;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.category-btn:hover {
  border-color: #0f3460;
  color: #0f3460;
}

.category-btn.active {
  background: #0f3460;
  border-color: #0f3460;
  color: #fff;
}
</style>
