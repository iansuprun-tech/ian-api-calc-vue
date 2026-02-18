# Clean Architecture — Полный конспект

## Оглавление

1. [Что такое архитектура и зачем она нужна](#1-что-такое-архитектура-и-зачем-она-нужна)
2. [Что такое Clean Architecture](#2-что-такое-clean-architecture)
3. [Главное правило — Dependency Rule](#3-главное-правило--dependency-rule)
4. [Слои Clean Architecture](#4-слои-clean-architecture)
5. [Как слои общаются между собой](#5-как-слои-общаются-между-собой)
6. [Пример: как выглядит код БЕЗ Clean Architecture](#6-пример-как-выглядит-код-без-clean-architecture)
7. [Пример: как выглядит код С Clean Architecture](#7-пример-как-выглядит-код-с-clean-architecture)
8. [Структура папок на Go (бэкенд)](#8-структура-папок-на-go-бэкенд)
9. [Структура папок на Vue/TS (фронтенд)](#9-структура-папок-на-vuets-фронтенд)
10. [SOLID — принципы, на которых стоит Clean Architecture](#10-solid--принципы-на-которых-стоит-clean-architecture)
11. [Частые вопросы новичков](#11-частые-вопросы-новичков)
12. [Итого: шпаргалка](#12-итого-шпаргалка)

---

## 1. Что такое архитектура и зачем она нужна

**Архитектура** — это то, как ты организуешь свой код. Куда какой файл положить,
какой код за что отвечает, кто кого вызывает.

### Проблема без архитектуры

Представь, что ты пишешь весь код в одном файле. Сначала всё ок — файл маленький.
Но проект растёт, и через месяц:

- Файл на 1000+ строк, невозможно найти нужное место
- Ты меняешь одну функцию — ломается три других
- Тесты писать невозможно, потому что всё связано со всем
- Другой разработчик не может понять, что происходит
- Хочешь поменять базу данных — нужно переписать половину кода

> **Архитектура решает все эти проблемы.** Она разделяет код на части (слои),
> у каждой части — своя ответственность, и менять одну часть можно
> без поломки остальных.

---

## 2. Что такое Clean Architecture

**Clean Architecture** (Чистая Архитектура) — это подход к организации кода,
придуманный Робертом Мартином (Uncle Bob) в 2012 году.

### Главная идея

> **Бизнес-логика (то, ЧТО делает программа) не должна зависеть от деталей
> (КАК она это делает).**

Что это значит на практике:

| Бизнес-логика (ЧТО)                      | Детали (КАК)                              |
|-------------------------------------------|-------------------------------------------|
| «Создать баланс с валютой и суммой»       | PostgreSQL, MySQL, файл, память           |
| «Получить курсы обмена»                   | HTTP-запрос к API, кеш, файл              |
| «Посчитать общую сумму в USD»             | Формула расчёта                           |
| «Показать список балансов пользователю»   | Vue, React, консоль, мобилка              |

Бизнес-логика — это **самое ценное** в программе. Она не меняется, когда ты
решаешь перейти с PostgreSQL на MongoDB или с Vue на React.

### Зачем нужна Clean Architecture

| Проблема                          | Как решает Clean Architecture                      |
|-----------------------------------|---------------------------------------------------|
| Код трудно понять                 | Каждый слой — про одно. Легко найти нужное         |
| Трудно менять БД/фреймворк        | Детали изолированы, меняешь только один слой       |
| Невозможно писать тесты           | Бизнес-логику можно тестировать без БД и сети      |
| Код ломается при изменениях       | Слои независимы, изменение одного не ломает другие |
| Новый человек не может разобраться | Чёткая структура, понятные правила                 |

---

## 3. Главное правило — Dependency Rule

Это **самое важное** правило Clean Architecture. Без него ничего не работает.

```
┌─────────────────────────────────────────────────┐
│                  Frameworks & DB                │  ← Внешний слой
│  ┌─────────────────────────────────────────┐    │
│  │           Controllers / Handlers        │    │  ← Адаптеры
│  │  ┌─────────────────────────────────┐    │    │
│  │  │          Use Cases              │    │    │  ← Бизнес-логика
│  │  │  ┌─────────────────────────┐    │    │    │
│  │  │  │       Entities          │    │    │    │  ← Ядро
│  │  │  └─────────────────────────┘    │    │    │
│  │  └─────────────────────────────────┘    │    │
│  └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

### Правило зависимостей (Dependency Rule)

> **Зависимости всегда направлены ВНУТРЬ. Внутренние слои НЕ ЗНАЮТ
> о внешних слоях.**

Что это значит:

- `Entities` не знают ни о чём вообще
- `Use Cases` знают только об `Entities`
- `Controllers` знают о `Use Cases` и `Entities`
- `Frameworks & DB` знают обо всём

**Простая аналогия:** Представь матрёшку. Маленькая матрёшка внутри не знает,
какая матрёшка снаружи. Но большая снаружи точно знает, что внутри неё есть
маленькая.

### Почему именно внутрь?

Потому что внутренние слои — это **стабильные** вещи (бизнес-правила не меняются
часто), а внешние — **нестабильные** (базу данных, фреймворк, API можно поменять).

Стабильные вещи не должны зависеть от нестабильных.

---

## 4. Слои Clean Architecture

### Слой 1: Entities (Сущности) — самый внутренний

**Что это:** Объекты, которые описывают данные и бизнес-правила предметной области.

**Простым языком:** Это «существительные» твоей программы. В калькуляторе валют —
это «Баланс» и «Курс обмена».

```go
// internal/entity/balance.go

package entity

// Balance — это сущность. Она описывает, ЧТО такое баланс.
// Она ничего не знает о базе данных, HTTP, Vue и т.д.
type Balance struct {
    ID       int
    Currency string
    Amount   float64
}

// Validate — бизнес-правило сущности.
// Сущность сама знает, что является допустимым балансом.
func (b Balance) Validate() error {
    if b.Currency == "" {
        return errors.New("валюта не может быть пустой")
    }
    if b.Amount < 0 {
        return errors.New("сумма не может быть отрицательной")
    }
    return nil
}
```

```go
// internal/entity/rate.go

package entity

type Rate struct {
    Currency string
    Rate     float64
}
```

**Ключевые свойства:**
- Не импортирует ничего из других слоёв
- Содержит только данные и правила, связанные с этими данными
- Самый стабильный слой — меняется очень редко

---

### Слой 2: Use Cases (Сценарии использования)

**Что это:** Бизнес-логика приложения. Описывает, ЧТО приложение ДЕЛАЕТ.

**Простым языком:** Это «глаголы» твоей программы — «создать баланс»,
«получить все балансы», «обновить курсы».

```go
// internal/usecase/balance.go

package usecase

import "vue-calc/internal/entity"

// BalanceRepository — интерфейс (контракт).
// Use Case ГОВОРИТ: «Мне нужен кто-то, кто умеет сохранять/читать балансы.
// Мне всё равно КАК — через БД, файл или память.»
type BalanceRepository interface {
    GetAll() ([]entity.Balance, error)
    GetByID(id int) (entity.Balance, error)
    Create(balance entity.Balance) (entity.Balance, error)
    Update(id int, balance entity.Balance) error
    Delete(id int) error
}

// BalanceUseCase — содержит бизнес-логику работы с балансами.
type BalanceUseCase struct {
    repo BalanceRepository  // зависимость через интерфейс!
}

// New — конструктор. Принимает любую реализацию BalanceRepository.
func New(repo BalanceRepository) *BalanceUseCase {
    return &BalanceUseCase{repo: repo}
}

// CreateBalance — сценарий «создать баланс».
func (uc *BalanceUseCase) CreateBalance(currency string, amount float64) (entity.Balance, error) {
    balance := entity.Balance{
        Currency: currency,
        Amount:   amount,
    }

    // 1. Проверяем бизнес-правила
    if err := balance.Validate(); err != nil {
        return entity.Balance{}, err
    }

    // 2. Сохраняем (нам всё равно КУДА — в БД, файл и т.д.)
    created, err := uc.repo.Create(balance)
    if err != nil {
        return entity.Balance{}, err
    }

    return created, nil
}

// GetAllBalances — сценарий «получить все балансы».
func (uc *BalanceUseCase) GetAllBalances() ([]entity.Balance, error) {
    return uc.repo.GetAll()
}
```

**Ключевые свойства:**
- Знает только о `Entities`
- Работает через **интерфейсы** (не знает конкретной реализации)
- Описывает БИЗНЕС-ЛОГИКУ, не технические детали
- Можно тестировать с фейковой (mock) реализацией репозитория

---

### Слой 3: Interface Adapters (Адаптеры интерфейсов)

**Что это:** Код, который преобразует данные между форматом Use Cases и форматом
внешнего мира (HTTP, БД и т.д.).

**Простым языком:** Это «переводчики». Они переводят HTTP-запрос в вызов Use Case,
а ответ Use Case — обратно в HTTP-ответ.

Сюда входят:
- **Handlers/Controllers** — обрабатывают HTTP-запросы
- **Repositories** — реализация работы с БД
- **Presenters** — форматируют данные для ответа

```go
// internal/handler/balance.go  (Controller/Handler)

package handler

import (
    "encoding/json"
    "net/http"
    "vue-calc/internal/usecase"
)

type BalanceHandler struct {
    useCase *usecase.BalanceUseCase
}

func NewBalanceHandler(uc *usecase.BalanceUseCase) *BalanceHandler {
    return &BalanceHandler{useCase: uc}
}

// GetAll обрабатывает GET /api/balances
// Его работа — принять HTTP-запрос, вызвать Use Case, отдать HTTP-ответ.
// Никакой бизнес-логики здесь НЕТ.
func (h *BalanceHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    // 1. Вызываем Use Case (бизнес-логику)
    balances, err := h.useCase.GetAllBalances()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Преобразуем результат в JSON и отправляем
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(balances)
}
```

```go
// internal/repository/postgres/balance.go  (Repository — реализация)

package postgres

import (
    "database/sql"
    "vue-calc/internal/entity"
)

// BalanceRepo реализует интерфейс usecase.BalanceRepository.
// Этот слой ЗНАЕТ про PostgreSQL. Но Use Case об этом не знает!
type BalanceRepo struct {
    db *sql.DB
}

func NewBalanceRepo(db *sql.DB) *BalanceRepo {
    return &BalanceRepo{db: db}
}

func (r *BalanceRepo) GetAll() ([]entity.Balance, error) {
    rows, err := r.db.Query("SELECT id, currency, amount FROM balances")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var balances []entity.Balance
    for rows.Next() {
        var b entity.Balance
        if err := rows.Scan(&b.ID, &b.Currency, &b.Amount); err != nil {
            return nil, err
        }
        balances = append(balances, b)
    }
    return balances, nil
}

func (r *BalanceRepo) Create(balance entity.Balance) (entity.Balance, error) {
    err := r.db.QueryRow(
        "INSERT INTO balances (currency, amount) VALUES ($1, $2) RETURNING id",
        balance.Currency, balance.Amount,
    ).Scan(&balance.ID)
    return balance, err
}

// ... остальные методы (GetByID, Update, Delete)
```

---

### Слой 4: Frameworks & Drivers (Фреймворки и драйверы) — самый внешний

**Что это:** Конкретные инструменты и библиотеки — база данных, веб-фреймворк,
внешние API.

**Простым языком:** Это «клей», который соединяет всё вместе. Здесь ты создаёшь
подключение к БД, настраиваешь роутер, запускаешь сервер.

```go
// cmd/app/main.go

package main

import (
    "database/sql"
    "log"
    "net/http"

    _ "github.com/lib/pq"

    "vue-calc/internal/handler"
    "vue-calc/internal/repository/postgres"
    "vue-calc/internal/usecase"
)

func main() {
    // 1. Инициализируем ВНЕШНИЕ зависимости (БД)
    db, err := sql.Open("postgres", "postgres://...")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Создаём РЕПОЗИТОРИЙ (слой адаптеров)
    //    Он реализует интерфейс, который определён в Use Case
    balanceRepo := postgres.NewBalanceRepo(db)

    // 3. Создаём USE CASE (бизнес-логика)
    //    Передаём ему репозиторий через интерфейс
    balanceUC := usecase.New(balanceRepo)

    // 4. Создаём HANDLER (слой адаптеров)
    //    Передаём ему Use Case
    balanceHandler := handler.NewBalanceHandler(balanceUC)

    // 5. Настраиваем роутер (внешний слой)
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/balances", balanceHandler.GetAll)

    // 6. Запускаем сервер
    log.Println("Сервер запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

**Обрати внимание на порядок создания:**
```
БД → Репозиторий → Use Case → Handler → Роутер → Сервер
     (внешний)                          (внешний)
              ↘   (внутренний)  ↙
```

Зависимости передаются **снаружи внутрь** при создании (это называется
**Dependency Injection** — внедрение зависимостей).

---

## 5. Как слои общаются между собой

```
HTTP-запрос
    │
    ▼
┌─────────────┐
│   Handler    │  Принимает HTTP-запрос, вытаскивает данные,
│ (адаптер)    │  вызывает Use Case
└──────┬──────┘
       │ вызывает
       ▼
┌─────────────┐
│  Use Case   │  Выполняет бизнес-логику,
│ (логика)     │  вызывает Repository через ИНТЕРФЕЙС
└──────┬──────┘
       │ вызывает через интерфейс
       ▼
┌─────────────┐
│ Repository  │  Выполняет SQL-запрос к PostgreSQL,
│ (адаптер)    │  возвращает Entity
└──────┬──────┘
       │
       ▼
  [PostgreSQL]

Ответ идёт в обратном порядке:
PostgreSQL → Repository → Use Case → Handler → HTTP-ответ
```

### Почему Repository — это интерфейс?

Это ключевой трюк Clean Architecture. Смотри:

```go
// Use Case определяет ИНТЕРФЕЙС (контракт):
type BalanceRepository interface {
    GetAll() ([]entity.Balance, error)
}

// PostgreSQL-реализация:
type PostgresBalanceRepo struct { db *sql.DB }
func (r *PostgresBalanceRepo) GetAll() ([]entity.Balance, error) {
    // SQL-запрос к PostgreSQL
}

// Для тестов — фейковая реализация:
type FakeBalanceRepo struct { data []entity.Balance }
func (r *FakeBalanceRepo) GetAll() ([]entity.Balance, error) {
    return r.data, nil  // просто возвращает данные из памяти
}
```

Оба типа реализуют один интерфейс. Use Case работает с интерфейсом, ему всё
равно, что за ним — настоящая БД или фейк. Это даёт:

1. **Тестируемость** — можно тестировать бизнес-логику без БД
2. **Заменяемость** — можно поменять PostgreSQL на MongoDB, изменив только один файл
3. **Независимость** — бизнес-логика не знает про SQL

---

## 6. Пример: как выглядит код БЕЗ Clean Architecture

Вот так сейчас выглядит типичный код без архитектуры (похоже на текущий `Main.go`):

```go
// Всё в одном файле, всё смешано

var db *sql.DB  // глобальная переменная — плохо!

func getBalances(w http.ResponseWriter, r *http.Request) {
    // HTTP-логика, SQL-запрос и бизнес-логика — всё в одной функции!

    rows, err := db.Query("SELECT id, currency, amount FROM balances")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()

    var balances []map[string]interface{}
    for rows.Next() {
        var id int
        var currency string
        var amount float64
        rows.Scan(&id, &currency, &amount)
        balances = append(balances, map[string]interface{}{
            "id": id, "currency": currency, "amount": amount,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(balances)
}
```

### Что плохо:

| Проблема                              | Почему плохо                                        |
|---------------------------------------|-----------------------------------------------------|
| Глобальная переменная `db`            | Невозможно тестировать, непонятно кто её использует  |
| SQL прямо в хендлере                  | Нельзя поменять БД без переписывания хендлеров      |
| Нет типизации (`map[string]interface{}`) | Нет проверки на этапе компиляции               |
| Всё в одном файле                     | Один файл на 1000+ строк — невозможно читать        |
| Нет валидации                         | Бизнес-правила не вынесены, их легко забыть          |
| Нельзя тестировать без БД             | Для теста нужна реальная PostgreSQL                  |

---

## 7. Пример: как выглядит код С Clean Architecture

Тот же функционал, но разделённый по слоям:

```
Запрос: GET /api/balances

1. Handler получает HTTP-запрос
   → Не знает про SQL, не знает бизнес-правила
   → Просто вызывает Use Case

2. Use Case выполняет бизнес-логику
   → Не знает про HTTP, не знает про SQL
   → Вызывает Repository через интерфейс

3. Repository делает SQL-запрос
   → Не знает про HTTP, не знает бизнес-правила
   → Возвращает Entity

4. Данные возвращаются обратно:
   Repository → Use Case → Handler → HTTP-ответ
```

**Каждый слой знает только одну вещь и делает только одну работу.**

---

## 8. Структура папок на Go (бэкенд)

Стандартная структура Go-проекта с Clean Architecture:

```
vue-calc/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа. Собирает всё вместе.
│
├── internal/                    # Внутренний код приложения
│   ├── entity/                  # Слой 1: Сущности
│   │   ├── balance.go           # type Balance struct { ... }
│   │   └── rate.go              # type Rate struct { ... }
│   │
│   ├── usecase/                 # Слой 2: Бизнес-логика
│   │   ├── balance.go           # CreateBalance, GetAllBalances, ...
│   │   └── rate.go              # GetRates, UpdateRates, ...
│   │
│   ├── handler/                 # Слой 3: HTTP-хендлеры (адаптеры)
│   │   ├── balance.go           # HTTP → Use Case → HTTP
│   │   └── rate.go
│   │
│   └── repository/              # Слой 3: Реализации репозиториев (адаптеры)
│       └── postgres/
│           ├── balance.go       # SQL-запросы для балансов
│           └── rate.go          # SQL-запросы для курсов
│
├── db/
│   └── migrations/              # SQL-миграции
│
├── go.mod
└── go.sum
```

### Что где лежит:

| Папка                     | Слой CA              | Что делает                            |
|---------------------------|----------------------|---------------------------------------|
| `internal/entity/`        | Entities             | Структуры данных + валидация          |
| `internal/usecase/`       | Use Cases            | Бизнес-логика + интерфейсы            |
| `internal/handler/`       | Interface Adapters   | HTTP-запросы/ответы                   |
| `internal/repository/`    | Interface Adapters   | SQL-запросы                           |
| `cmd/app/`                | Frameworks & Drivers | Сборка + запуск                       |

---

## 9. Структура папок на Vue/TS (фронтенд)

Clean Architecture можно применить и на фронтенде:

```
src/
├── entities/                    # Слой 1: Типы и бизнес-правила
│   ├── balance.ts               # interface Balance { id: number; ... }
│   └── rate.ts                  # interface Rate { currency: string; ... }
│
├── usecases/                    # Слой 2: Бизнес-логика
│   └── calculateTotal.ts        # Функция расчёта общей суммы в USD
│
├── services/                    # Слой 3: Адаптеры для внешних API
│   └── api.ts                   # Все fetch-запросы к бэкенду
│
├── stores/                      # Слой 3: Управление состоянием
│   ├── balanceStore.ts          # Pinia store для балансов
│   └── rateStore.ts             # Pinia store для курсов
│
├── components/                  # Слой 4: UI-компоненты
│   ├── BalanceList.vue
│   ├── BalanceForm.vue
│   └── RateBanner.vue
│
├── views/                       # Слой 4: Страницы
│   └── CalculatorView.vue
│
├── router/
│   └── index.ts
│
├── App.vue
└── main.ts
```

### Пример — выделение API-сервиса:

**Было (всё внутри компонента):**
```vue
<script setup>
// Прямо в компоненте — fetch-запросы, бизнес-логика, отображение
const response = await fetch('/api/balances')
const balances = await response.json()

const total = balances.reduce((sum, b) => {
  const rate = rates.find(r => r.currency === b.currency)
  return sum + b.amount * (rate?.rate || 0)
}, 0)
</script>
```

**Стало (разделено по слоям):**

```typescript
// entities/balance.ts — Слой 1
export interface Balance {
  id: number
  currency: string
  amount: number
}
```

```typescript
// services/api.ts — Слой 3 (адаптер)
import type { Balance } from '@/entities/balance'

export async function fetchBalances(): Promise<Balance[]> {
  const response = await fetch('/api/balances')
  return response.json()
}

export async function createBalance(currency: string, amount: number): Promise<Balance> {
  const response = await fetch('/api/balances', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ currency, amount }),
  })
  return response.json()
}
```

```typescript
// usecases/calculateTotal.ts — Слой 2 (бизнес-логика)
import type { Balance } from '@/entities/balance'
import type { Rate } from '@/entities/rate'

export function calculateTotalUSD(balances: Balance[], rates: Rate[]): number {
  return balances.reduce((sum, balance) => {
    const rate = rates.find(r => r.currency === balance.currency)
    if (!rate) return sum
    return sum + balance.amount * rate.rate
  }, 0)
}
```

```vue
<!-- views/CalculatorView.vue — Слой 4 (UI) -->
<script setup>
import { fetchBalances } from '@/services/api'
import { calculateTotalUSD } from '@/usecases/calculateTotal'

// Компонент только ОТОБРАЖАЕТ, логика — в других слоях
const balances = await fetchBalances()
const total = calculateTotalUSD(balances, rates)
</script>
```

---

## 10. SOLID — принципы, на которых стоит Clean Architecture

Clean Architecture построена на принципах SOLID. Вот они простым языком:

### S — Single Responsibility (Единственная ответственность)

> Каждый модуль/класс/функция отвечает за ОДНУ вещь.

```
Handler    → отвечает только за HTTP
Use Case   → отвечает только за бизнес-логику
Repository → отвечает только за работу с БД
```

**Плохо:** Функция, которая и парсит HTTP, и ходит в БД, и считает формулы.
**Хорошо:** Три отдельные функции, каждая делает своё.

### O — Open/Closed (Открытость/Закрытость)

> Код открыт для расширения, закрыт для модификации.

Хочешь добавить новый тип хранилища (Redis)? Не меняешь Use Case — просто
пишешь новую реализацию интерфейса `BalanceRepository`.

### L — Liskov Substitution (Подстановка Лисков)

> Любая реализация интерфейса должна быть взаимозаменяемой.

`PostgresBalanceRepo` и `FakeBalanceRepo` оба реализуют `BalanceRepository`.
Use Case работает одинаково с любым из них.

### I — Interface Segregation (Разделение интерфейсов)

> Лучше много маленьких интерфейсов, чем один большой.

```go
// Плохо: один огромный интерфейс
type Storage interface {
    GetBalance(id int) (Balance, error)
    CreateBalance(b Balance) error
    GetRate(currency string) (Rate, error)
    SendEmail(to string, body string) error  // зачем здесь email?!
}

// Хорошо: маленькие, специализированные интерфейсы
type BalanceRepository interface {
    GetAll() ([]Balance, error)
    Create(b Balance) (Balance, error)
}

type RateRepository interface {
    GetAll() ([]Rate, error)
    Update(rates []Rate) error
}
```

### D — Dependency Inversion (Инверсия зависимостей)

> Зависимость направлена от деталей к абстракциям, а не наоборот.

Это основа Clean Architecture:

```go
// Use Case зависит от ИНТЕРФЕЙСА (абстракции)
type BalanceUseCase struct {
    repo BalanceRepository  // интерфейс, не конкретный тип!
}

// А конкретная реализация (PostgreSQL) зависит от этого же интерфейса
type PostgresRepo struct { ... }
func (r *PostgresRepo) GetAll() ([]Balance, error) { ... }
```

**Без инверсии:** Use Case → PostgresRepo (жёсткая связь)
**С инверсией:** Use Case → Interface ← PostgresRepo (слабая связь)

---

## 11. Частые вопросы новичков

### «Это слишком много файлов для маленького проекта!»

Да, для проекта на 300 строк Clean Architecture — оверинжиниринг. Она начинает
приносить пользу, когда проект растёт. Но учиться на маленьком проекте — это
нормально и правильно.

Правило: если у тебя **один разработчик и проект на пару недель** — можно
обойтись без строгой архитектуры. Если **проект будет жить долго или работает
команда** — архитектура сэкономит кучу времени.

### «Зачем интерфейсы, если я точно буду использовать PostgreSQL?»

1. **Тесты.** Без интерфейсов ты не сможешь тестировать бизнес-логику
   без реальной БД.
2. **Будущее.** Может, завтра тебе нужно будет кешировать в Redis.
   С интерфейсом — это один новый файл.
3. **Читаемость.** Интерфейс — это документация: «вот что мне нужно от
   хранилища».

### «Где должна быть валидация?»

- **Валидация бизнес-правил** → в Entity (`Amount >= 0`, `Currency не пустой`)
- **Валидация формата входных данных** → в Handler (`JSON корректный?`, `ID число?`)
- **Валидация бизнес-логики** → в Use Case (`Баланс с такой валютой уже существует?`)

### «Чем отличается от MVC?»

| Аспект          | MVC                               | Clean Architecture                          |
|-----------------|-----------------------------------|---------------------------------------------|
| Слои            | 3 (Model-View-Controller)         | 4+ (Entity, UseCase, Adapter, Framework)    |
| Зависимости     | Нет строгих правил                | Строгое правило: только внутрь              |
| Бизнес-логика   | Часто в Controller или Model      | Строго в Use Case                           |
| Тестируемость   | Средняя                           | Высокая (каждый слой тестируется отдельно)  |
| Сложность       | Низкая                            | Выше, но окупается на больших проектах      |

### «Можно ли нарушить правила?»

Да, Clean Architecture — это **руководство**, а не закон. В реальных проектах
бывают компромиссы. Главное — понимать, ЗАЧЕМ ты нарушаешь правило и ЧЕМ
расплачиваешься.

---

## 12. Итого: шпаргалка

### Слои (снизу вверх):

```
┌──────────────────────────────────────────────┐
│ 4. Frameworks    │ main.go, роутер, БД-драйв │ ← Меняется часто
├──────────────────────────────────────────────┤
│ 3. Adapters      │ Handlers, Repositories     │ ← Преобразует данные
├──────────────────────────────────────────────┤
│ 2. Use Cases     │ Бизнес-логика + интерфейсы │ ← Что делает приложение
├──────────────────────────────────────────────┤
│ 1. Entities      │ Структуры + правила        │ ← Почти никогда не меняется
└──────────────────────────────────────────────┘
```

### Главные правила:

1. **Dependency Rule** — зависимости только ВНУТРЬ (внутренний слой НЕ знает
   о внешнем)
2. **Интерфейсы** — Use Case определяет интерфейс, внешний слой реализует его
3. **Dependency Injection** — зависимости передаются при создании (через конструктор)
4. **Каждый слой — одна ответственность**

### Что в каком слое:

| Вопрос                                     | Ответ               |
|--------------------------------------------|----------------------|
| Куда положить `type Balance struct`?       | Entity               |
| Куда положить `CreateBalance()`?           | Use Case             |
| Куда положить SQL-запрос?                  | Repository (Adapter) |
| Куда положить парсинг HTTP-запроса?        | Handler (Adapter)    |
| Куда положить подключение к БД?            | main.go (Framework)  |
| Куда положить интерфейс `BalanceRepository`? | Use Case           |

### Одной фразой:

> **Clean Architecture = бизнес-логика в центре, детали снаружи, зависимости
> внутрь, общение через интерфейсы.**

---

*Конспект подготовлен на основе книги «Clean Architecture» Роберта Мартина (2017)
и адаптирован под проект vue-calc (Go + Vue).*
