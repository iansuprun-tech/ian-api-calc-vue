<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { apiFetch } from '@/api'
import { formatAmount } from '@/format'
import { Pie, Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
} from 'chart.js'

ChartJS.register(ArcElement, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

type CategoryStat = {
  category_id: number | null
  category_name: string
  total: number
  count: number
}

type DailyStat = {
  date: string
  income: number
  expense: number
}

type CurrencyStats = {
  currency: string
  total_income: number
  total_expense: number
  income_by_category: CategoryStat[]
  expense_by_category: CategoryStat[]
  daily_stats: DailyStat[]
}

type StatisticsData = {
  currencies: CurrencyStats[]
}

type Account = {
  id: number
  currency: string
  comment: string
  balance: number
}

type PeriodType = 'day' | 'week' | 'month' | 'year' | 'custom'

const periods: { key: PeriodType; label: string }[] = [
  { key: 'day', label: 'День' },
  { key: 'week', label: 'Неделя' },
  { key: 'month', label: 'Месяц' },
  { key: 'year', label: 'Год' },
  { key: 'custom', label: 'Период' },
]

const activePeriod = ref<PeriodType>('month')
const dateFrom = ref('')
const dateTo = ref('')
const selectedAccountId = ref<number | null>(null)
const accounts = ref<Account[]>([])
const stats = ref<StatisticsData | null>(null)
const loading = ref(false)

function toLocalDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function setDatesForPeriod(period: PeriodType) {
  const now = new Date()
  const to = toLocalDate(now)
  let from = to

  switch (period) {
    case 'day':
      from = to
      break
    case 'week': {
      const d = new Date(now)
      d.setDate(d.getDate() - 6)
      from = toLocalDate(d)
      break
    }
    case 'month': {
      const d = new Date(now.getFullYear(), now.getMonth(), 1)
      from = toLocalDate(d)
      break
    }
    case 'year': {
      const d = new Date(now.getFullYear(), 0, 1)
      from = toLocalDate(d)
      break
    }
    case 'custom':
      return
  }

  dateFrom.value = from
  dateTo.value = to
}

function selectPeriod(period: PeriodType) {
  activePeriod.value = period
  setDatesForPeriod(period)
}

async function loadAccounts() {
  const response = await apiFetch('/api/accounts')
  if (response.ok) {
    accounts.value = await response.json()
  }
}

async function loadStats() {
  if (!dateFrom.value || !dateTo.value) return
  loading.value = true

  let url = `/api/statistics?from=${dateFrom.value}&to=${dateTo.value}`
  if (selectedAccountId.value !== null) {
    url += `&account_id=${selectedAccountId.value}`
  }

  const response = await apiFetch(url)
  if (response.ok) {
    stats.value = await response.json()
  }
  loading.value = false
}

watch([dateFrom, dateTo, selectedAccountId], () => {
  loadStats()
})

onMounted(() => {
  setDatesForPeriod('month')
  loadAccounts()
  loadStats()
})

const pieColors = [
  '#0f3460', '#e94560', '#16213e', '#0ea5e9', '#22c55e',
  '#f59e0b', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316',
  '#6366f1', '#84cc16', '#ef4444', '#06b6d4', '#a855f7',
]

function makeExpensePieData(cs: CurrencyStats) {
  if (cs.expense_by_category.length === 0) return null
  return {
    labels: cs.expense_by_category.map((c) => c.category_name),
    datasets: [
      {
        data: cs.expense_by_category.map((c) => c.total),
        backgroundColor: pieColors.slice(0, cs.expense_by_category.length),
      },
    ],
  }
}

function makeIncomePieData(cs: CurrencyStats) {
  if (cs.income_by_category.length === 0) return null
  return {
    labels: cs.income_by_category.map((c) => c.category_name),
    datasets: [
      {
        data: cs.income_by_category.map((c) => c.total),
        backgroundColor: pieColors.slice(0, cs.income_by_category.length),
      },
    ],
  }
}

function makeBarData(cs: CurrencyStats) {
  if (cs.daily_stats.length === 0) return null
  return {
    labels: cs.daily_stats.map((d) => d.date.slice(5)),
    datasets: [
      {
        label: 'Доходы',
        data: cs.daily_stats.map((d) => d.income),
        backgroundColor: '#22c55e',
      },
      {
        label: 'Расходы',
        data: cs.daily_stats.map((d) => d.expense),
        backgroundColor: '#e94560',
      },
    ],
  }
}

const pieOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: { padding: 12, usePointStyle: true },
    },
  },
}

const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      labels: { usePointStyle: true },
    },
  },
  scales: {
    x: { grid: { display: false } },
    y: { beginAtZero: true },
  },
}

const hasData = computed(() => {
  return stats.value && stats.value.currencies.length > 0
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Статистика</h1>
    </div>

    <!-- Управление -->
    <div class="card controls-card">
      <div class="controls-row">
        <div class="period-buttons">
          <button
            v-for="p in periods"
            :key="p.key"
            class="period-btn"
            :class="{ active: activePeriod === p.key }"
            @click="selectPeriod(p.key)"
          >
            {{ p.label }}
          </button>
        </div>

        <div class="date-inputs">
          <input type="date" v-model="dateFrom" class="input-field" />
          <span class="date-sep">&mdash;</span>
          <input type="date" v-model="dateTo" class="input-field" />
        </div>

        <select v-model="selectedAccountId" class="input-field select-field">
          <option :value="null">Все счета</option>
          <option v-for="a in accounts" :key="a.id" :value="a.id">
            {{ a.comment || a.currency }} ({{ a.currency }})
          </option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="loading">Загрузка...</div>

    <template v-if="stats && !loading">
      <template v-if="hasData">
        <div
          v-for="cs in stats.currencies"
          :key="cs.currency"
          class="currency-section"
        >
          <h2 class="currency-title">{{ cs.currency }}</h2>

          <!-- Карточки-итоги -->
          <div class="summary-row">
            <div class="summary-card income">
              <div class="summary-label">Доходы</div>
              <div class="summary-value">{{ formatAmount(cs.total_income) }} {{ cs.currency }}</div>
            </div>
            <div class="summary-card expense">
              <div class="summary-label">Расходы</div>
              <div class="summary-value">{{ formatAmount(cs.total_expense) }} {{ cs.currency }}</div>
            </div>
            <div
              class="summary-card net"
              :class="{
                positive: cs.total_income - cs.total_expense >= 0,
                negative: cs.total_income - cs.total_expense < 0,
              }"
            >
              <div class="summary-label">Разница</div>
              <div class="summary-value">
                {{ formatAmount(cs.total_income - cs.total_expense) }} {{ cs.currency }}
              </div>
            </div>
          </div>

          <!-- Pie charts -->
          <div
            class="charts-grid"
            v-if="makeExpensePieData(cs) || makeIncomePieData(cs)"
          >
            <div class="card chart-card" v-if="makeExpensePieData(cs)">
              <h3 class="card-title">Расходы по категориям</h3>
              <div class="chart-container">
                <Pie :data="makeExpensePieData(cs)!" :options="pieOptions" />
              </div>
            </div>

            <div class="card chart-card" v-if="makeIncomePieData(cs)">
              <h3 class="card-title">Доходы по категориям</h3>
              <div class="chart-container">
                <Pie :data="makeIncomePieData(cs)!" :options="pieOptions" />
              </div>
            </div>
          </div>

          <!-- Bar chart -->
          <div class="card" v-if="makeBarData(cs)">
            <h3 class="card-title">По дням</h3>
            <div class="bar-container">
              <Bar :data="makeBarData(cs)!" :options="barOptions" />
            </div>
          </div>
        </div>
      </template>

      <div v-else class="card empty-state">
        <p>Нет данных за выбранный период</p>
      </div>
    </template>
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

.card {
  background: #fff;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.card-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 1rem;
}

.controls-card {
  margin-bottom: 1.5rem;
}

.controls-row {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.period-buttons {
  display: flex;
  gap: 0.25rem;
}

.period-btn {
  padding: 0.45rem 0.9rem;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  background: #fff;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  color: #555;
}

.period-btn:hover {
  border-color: #0f3460;
  color: #0f3460;
}

.period-btn.active {
  background: #0f3460;
  color: #fff;
  border-color: #0f3460;
}

.date-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-sep {
  color: #999;
}

.input-field {
  padding: 0.5rem 0.75rem;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 0.9rem;
  background: #fff;
  color: #333;
  transition: border-color 0.2s;
}

.input-field:focus {
  outline: none;
  border-color: #0f3460;
}

.select-field {
  min-width: 160px;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #999;
  font-size: 1rem;
}

/* --- Currency section --- */
.currency-section {
  margin-bottom: 2.5rem;
}

.currency-title {
  font-size: 1.3rem;
  font-weight: 700;
  color: #0f3460;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #0f3460;
}

.summary-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.summary-card {
  background: #fff;
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.summary-label {
  font-size: 0.85rem;
  color: #888;
  margin-bottom: 0.4rem;
  font-weight: 500;
}

.summary-value {
  font-size: 1.3rem;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.summary-card.income .summary-value {
  color: #22c55e;
}

.summary-card.expense .summary-value {
  color: #e94560;
}

.summary-card.net.positive .summary-value {
  color: #22c55e;
}

.summary-card.net.negative .summary-value {
  color: #e94560;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.chart-card {
  min-height: 320px;
}

.chart-container {
  height: 260px;
  position: relative;
}

.bar-container {
  height: 300px;
  position: relative;
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #999;
}

@media (max-width: 768px) {
  .controls-row {
    flex-direction: column;
    align-items: stretch;
  }

  .period-buttons {
    flex-wrap: wrap;
  }

  .summary-row {
    grid-template-columns: 1fr;
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }
}
</style>
