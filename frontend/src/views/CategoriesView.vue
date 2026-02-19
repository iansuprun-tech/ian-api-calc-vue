<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiFetch } from '@/api'

type Category = {
  id: number
  name: string
  created_at: string
}

const categories = ref<Category[]>([])
const newName = ref('')
const deletingId = ref<number | null>(null)

function loadCategories() {
  apiFetch('/api/categories')
    .then((response) => response.json())
    .then((data) => (categories.value = data))
}

onMounted(() => {
  loadCategories()
})

async function addCategory() {
  const name = newName.value.trim()
  if (!name) return

  const response = await apiFetch('/api/categories', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
  if (response.ok) {
    newName.value = ''
    loadCategories()
  }
}

function confirmDelete(id: number) {
  deletingId.value = id
}

function cancelDelete() {
  deletingId.value = null
}

async function deleteCategory() {
  if (deletingId.value === null) return
  const response = await apiFetch(`/api/categories/${deletingId.value}`, {
    method: 'DELETE',
  })
  deletingId.value = null
  if (response.ok) {
    loadCategories()
  }
}
</script>

<template>
  <div class="page">
    <h1 class="page-title">Категории</h1>

    <div class="card">
      <h2 class="card-title">Добавить категорию</h2>
      <form @submit.prevent="addCategory" class="add-form">
        <input
          v-model="newName"
          placeholder="Название категории"
          class="input-field input-grow"
        />
        <button type="submit" class="btn btn-primary">Добавить</button>
      </form>
    </div>

    <div class="card">
      <h2 class="card-title">Мои категории</h2>

      <div v-if="categories.length === 0" class="empty-state">
        Категорий пока нет
      </div>

      <div class="category-list">
        <div v-for="cat in categories" :key="cat.id" class="category-chip">
          <span class="category-name">{{ cat.name }}</span>
          <button class="chip-delete" @click="confirmDelete(cat.id)">&times;</button>
        </div>
      </div>
    </div>

    <!-- Модалка подтверждения удаления -->
    <Teleport to="body">
      <div v-if="deletingId !== null" class="modal-overlay" @click.self="cancelDelete">
        <div class="modal-card">
          <p class="modal-text">Удалить категорию?</p>
          <p class="modal-subtext">Существующие операции сохранятся, но категория будет убрана.</p>
          <div class="modal-buttons">
            <button class="btn btn-outline" @click="cancelDelete">Отмена</button>
            <button class="btn btn-danger" @click="deleteCategory">Удалить</button>
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

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #1a1a2e;
  margin-bottom: 1.5rem;
}

.card {
  background: #fff;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  margin-bottom: 1.5rem;
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
}

.input-grow {
  flex: 1;
}

.input-field:focus {
  outline: none;
  border-color: #0f3460;
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

.btn-primary {
  background: #0f3460;
  color: #fff;
}

.btn-primary:hover {
  background: #0a2540;
}

.btn-danger {
  background: #d73a49;
  color: #fff;
}

.btn-danger:hover {
  background: #b42d3a;
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

.empty-state {
  text-align: center;
  padding: 2rem 1rem;
  color: #999;
}

.category-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.category-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: #f0f4f8;
  border: 1px solid #d0d7de;
  border-radius: 20px;
  padding: 0.4rem 0.6rem 0.4rem 0.9rem;
  font-size: 0.9rem;
  color: #333;
}

.category-name {
  font-weight: 500;
}

.chip-delete {
  background: none;
  border: none;
  color: #999;
  font-size: 1.1rem;
  line-height: 1;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.chip-delete:hover {
  background: #d73a49;
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
  margin-bottom: 0.5rem;
}

.modal-subtext {
  font-size: 0.85rem;
  color: #888;
  margin-bottom: 1.25rem;
}

.modal-buttons {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}
</style>
