# Что мы изменили и зачем — простым языком

---

## Было vs Стало (общая картина)

**Раньше** у нас была одна табличка `balances` — это как листочек, где написано:

```
| USD | 100 |
| EUR | 50  |
```

Просто валюта и сумма. Всё. Нельзя узнать откуда взялись деньги, нельзя завести два кошелька в одной валюте.

**Теперь** у нас две таблички — `accounts` (счета) и `transactions` (операции):

```
СЧЕТА:
| id | валюта | комментарий      |
|----|--------|------------------|
| 1  | USD    | Зарплатный       |
| 2  | USD    | Копилка          |
| 3  | EUR    | Путешествия      |

ОПЕРАЦИИ:
| id | счёт | сумма  | комментарий    |
|----|------|--------|----------------|
| 1  | 1    | +5000  | Зарплата       |
| 2  | 1    | -200   | Продукты       |
| 3  | 2    | +1000  | Отложил        |
| 4  | 3    | +500   | Подарок        |
```

Баланс счёта = сумма всех его операций. Например, счёт 1: `5000 + (-200) = 4800`.

---

## Архитектура проекта (как устроен код)

Представь себе пиццерию:

```
Клиент (браузер)
    ↓ делает заказ
Официант (handler) — принимает заказ, отдаёт ответ
    ↓ передаёт на кухню
Шеф-повар (usecase) — решает, как готовить
    ↓ берёт продукты
Холодильник (repository) — хранит и достаёт данные из БД
    ↓
База данных (PostgreSQL) — сами продукты
```

В коде это 4 папки:
- `handler/` — принимает HTTP-запросы от браузера
- `usecase/` — бизнес-логика (что делать с данными)
- `repository/` — запросы к базе данных (SQL)
- `entity/` — описание данных (как выглядит "счёт", "транзакция")

---

## Шаг 1: Миграции (новые таблицы в базе данных)

### Что такое миграция?

Миграция — это SQL-файл, который говорит базе данных: "Создай новую таблицу" или "Удали таблицу". Как инструкция для строителя.

### Файл `000003_create_accounts_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    currency TEXT NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Разберём по строчкам:

| Строка | Что делает |
|--------|-----------|
| `CREATE TABLE IF NOT EXISTS accounts` | Создай таблицу "accounts", если её ещё нет |
| `id SERIAL PRIMARY KEY` | Колонка "id" — автоматический номер (1, 2, 3...), уникальный ключ |
| `currency TEXT NOT NULL` | Колонка "валюта" — текст, обязательна (не может быть пустой) |
| `comment TEXT NOT NULL DEFAULT ''` | Колонка "комментарий" — текст, по умолчанию пустая строка |
| `created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP` | Колонка "дата создания" — автоматически ставит текущее время |

### Файл `000004_create_transactions_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount DOUBLE PRECISION NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Новое здесь:

| Строка | Что делает |
|--------|-----------|
| `account_id INTEGER NOT NULL REFERENCES accounts(id)` | Ссылка на счёт. Транзакция всегда привязана к конкретному счёту |
| `ON DELETE CASCADE` | Если удалим счёт — все его транзакции удалятся автоматически |
| `amount DOUBLE PRECISION NOT NULL` | Сумма — дробное число (может быть +100.50 или -30.00) |

**Аналогия:** `REFERENCES` — это как ярлык на папку. Транзакция "знает", к какому счёту относится. А `CASCADE` — если удалить папку, все файлы внутри тоже удалятся.

### down-файлы

```sql
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS transactions;
```

Это "откат" — если что-то пошло не так, удаляем таблицы. Как кнопка "Отменить".

---

## Шаг 2: Entity (модели данных)

### Что такое entity?

Entity — это описание "как выглядит штука". Как чертёж дома перед строительством.

### `internal/entity/account.go`

```go
type Account struct {
    ID        int     `json:"id"`
    Currency  string  `json:"currency"`
    Comment   string  `json:"comment"`
    CreatedAt string  `json:"created_at"`
    Balance   float64 `json:"balance"`
}
```

Разберём:

| Поле | Тип | Что это |
|------|-----|---------|
| `ID` | `int` | Номер счёта (1, 2, 3...) |
| `Currency` | `string` | Валюта ("USD", "EUR", "RUB") |
| `Comment` | `string` | Комментарий ("Зарплатный", "Копилка") |
| `CreatedAt` | `string` | Когда создан |
| `Balance` | `float64` | Баланс — **не хранится в БД!** Вычисляется как сумма транзакций |

**Что такое `` `json:"id"` ``?** Это подсказка для Go: "Когда отправляешь этот объект в браузер как JSON, называй поле `id`, а не `ID`". Браузер получит:

```json
{"id": 1, "currency": "USD", "comment": "Зарплатный", "balance": 4800}
```

### `internal/entity/transaction.go`

```go
type Transaction struct {
    ID        int     `json:"id"`
    AccountID int     `json:"account_id"`
    Amount    float64 `json:"amount"`
    Comment   string  `json:"comment"`
    CreatedAt string  `json:"created_at"`
}
```

| Поле | Что это | Пример |
|------|---------|--------|
| `AccountID` | К какому счёту относится | `1` |
| `Amount` | Сумма: `+` = пополнение, `-` = списание | `+5000` или `-200` |
| `Comment` | За что операция | "Зарплата", "Продукты" |

---

## Шаг 3: Repository (работа с базой данных)

### Что такое repository?

Repository — это "переводчик" между Go-кодом и SQL-запросами. Go не понимает SQL напрямую, а repository говорит: "Хочешь все счета? Сейчас спрошу у базы".

### `internal/repository/postgres/account.go`

#### Получить все счета (`GetAll`)

```go
func (r *AccountRepo) GetAll() ([]entity.Account, error) {
    rows, err := r.db.Query(`
        SELECT a.id, a.currency, a.comment, a.created_at,
               COALESCE(
                   (SELECT SUM(t.amount) FROM transactions t WHERE t.account_id = a.id),
                   0
               ) AS balance
        FROM accounts a
        ORDER BY a.id
    `)
    ...
}
```

Что тут происходит:

1. `SELECT ... FROM accounts a` — берём все счета
2. `(SELECT SUM(t.amount) FROM transactions t WHERE t.account_id = a.id)` — для каждого счёта считаем сумму всех его транзакций
3. `COALESCE(..., 0)` — если транзакций нет, вернуть 0 (а не NULL)

**Аналогия с COALESCE:** Ты спрашиваешь "Сколько у меня денег?". Если есть операции — считаем. Если операций ноль — отвечаем "0", а не "не знаю".

#### Создать счёт (`Create`)

```go
func (r *AccountRepo) Create(account entity.Account) (entity.Account, error) {
    err := r.db.QueryRow(
        "INSERT INTO accounts (currency, comment) VALUES ($1, $2) RETURNING id, created_at",
        account.Currency, account.Comment,
    ).Scan(&account.ID, &account.CreatedAt)
    return account, err
}
```

1. `INSERT INTO accounts (currency, comment) VALUES ($1, $2)` — вставляем новую строку
2. `$1, $2` — это "заглушки" для защиты от SQL-инъекций. Go подставит туда `account.Currency` и `account.Comment`
3. `RETURNING id, created_at` — база сама сгенерирует id и дату, и вернёт их нам
4. `.Scan(&account.ID, &account.CreatedAt)` — записываем возвращённые значения в наш объект

**Зачем `$1, $2` вместо прямой вставки?** Безопасность. Если пользователь введёт `'; DROP TABLE accounts; --` как комментарий, `$1` безопасно обработает это как обычный текст.

#### Удалить счёт (`Delete`)

```go
func (r *AccountRepo) Delete(id int) (int64, error) {
    result, err := r.db.Exec("DELETE FROM accounts WHERE id = $1", id)
    ...
    return result.RowsAffected()
}
```

`RowsAffected()` возвращает количество удалённых строк. Если 0 — значит такого счёта не было.

### `internal/repository/postgres/transaction.go`

#### Получить транзакции по счёту (`GetByAccountID`)

```go
rows, err := r.db.Query(
    "SELECT id, account_id, amount, comment, created_at FROM transactions WHERE account_id = $1 ORDER BY created_at DESC",
    accountID,
)
```

`ORDER BY created_at DESC` — сортировка от новых к старым. DESC = descending = по убыванию.

#### Создать транзакцию (`Create`)

```go
err := r.db.QueryRow(
    "INSERT INTO transactions (account_id, amount, comment) VALUES ($1, $2, $3) RETURNING id, created_at",
    transaction.AccountID, transaction.Amount, transaction.Comment,
).Scan(&transaction.ID, &transaction.CreatedAt)
```

Тот же паттерн: вставляем строку, получаем назад `id` и `created_at`.

---

## Шаг 4: UseCase (бизнес-логика)

### Что такое usecase?

UseCase — это "менеджер", который решает что делать. Сейчас он простой (просто вызывает repository), но если добавить правила (например, "нельзя списать больше, чем есть на счёте"), они будут именно тут.

### `internal/usecase/account.go`

```go
// Интерфейс — контракт: "Репозиторий должен уметь вот это"
type AccountRepository interface {
    GetAll() ([]entity.Account, error)
    GetByID(id int) (entity.Account, error)
    Create(account entity.Account) (entity.Account, error)
    Delete(id int) (int64, error)
    Exists(id int) (bool, error)
}

// UseCase знает только про интерфейс, а не про конкретную реализацию
type AccountUseCase struct {
    repo AccountRepository  // ← это интерфейс, не конкретный postgres-репозиторий
}
```

**Зачем интерфейс?** Представь, что `AccountRepository` — это розетка. Ты можешь воткнуть туда PostgreSQL, MySQL, или даже "фейковый" репозиторий для тестов. UseCase не знает и не заботится что внутри — главное, чтобы были нужные методы.

```
UseCase → [интерфейс AccountRepository] ← PostgresRepo
                                         ← MySQLRepo (в будущем)
                                         ← MockRepo (для тестов)
```

---

## Шаг 5: Handler (HTTP-обработчики)

### Что такое handler?

Handler — это "официант". Браузер посылает HTTP-запрос, handler его принимает, просит usecase выполнить работу, и возвращает ответ.

### `internal/handler/account.go`

#### Маршрутизация запросов

```go
func (h *AccountHandler) HandleList(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:    // GET /api/accounts → получить все счета
        h.getAll(w, r)
    case http.MethodPost:   // POST /api/accounts → создать счёт
        h.create(w, r)
    }
}
```

**Что такое `w` и `r`?**
- `r` (request) — входящий запрос от браузера. Содержит метод (GET/POST), URL, тело запроса
- `w` (writer) — через него мы пишем ответ обратно в браузер

**Аналогия:**
- `r` — записка от клиента: "Хочу пиццу Маргарита"
- `w` — поднос, на который кладём готовую пиццу

#### Создание счёта

```go
func (h *AccountHandler) create(w http.ResponseWriter, r *http.Request) {
    // 1. Читаем JSON из тела запроса
    var account entity.Account
    json.NewDecoder(r.Body).Decode(&account)

    // 2. Просим usecase создать счёт
    account, err := h.uc.Create(account)

    // 3. Отправляем ответ: статус 201 (Created) + созданный счёт как JSON
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(account)
}
```

Браузер отправляет:
```json
POST /api/accounts
{"currency": "USD", "comment": "Зарплатный"}
```

Сервер отвечает:
```json
201 Created
{"id": 1, "currency": "USD", "comment": "Зарплатный", "created_at": "2026-02-19...", "balance": 0}
```

### `internal/handler/transaction.go`

#### Парсинг URL

```go
// URL: /api/accounts/5/transactions
path := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
// path = "5/transactions"

parts := strings.SplitN(path, "/", 2)
// parts = ["5", "transactions"]

accountID, _ := strconv.Atoi(parts[0])
// accountID = 5
```

Пошагово:
1. Убираем `/api/accounts/` из начала → остаётся `5/transactions`
2. Разрезаем по `/` на 2 части → `["5", "transactions"]`
3. Первая часть — это ID счёта → преобразуем строку `"5"` в число `5`

---

## Шаг 6: main.go (точка входа)

### Что делает main.go?

Это "директор" — собирает всех вместе и запускает сервер.

```go
// 1. Создаём репозитории (подключаем к базе)
accountRepo := postgres.NewAccountRepo(db)
transactionRepo := postgres.NewTransactionRepo(db)

// 2. Создаём юзкейсы (даём им репозитории)
accountUC := usecase.NewAccountUseCase(accountRepo)
transactionUC := usecase.NewTransactionUseCase(transactionRepo)

// 3. Создаём хендлеры (даём им юзкейсы)
accountHandler := handler.NewAccountHandler(accountUC)
transactionHandler := handler.NewTransactionHandler(transactionUC, accountUC)
```

Это и есть **Dependency Injection** (внедрение зависимостей):

```
main.go создаёт accountRepo
    ↓ передаёт в
main.go создаёт accountUC(accountRepo)
    ↓ передаёт в
main.go создаёт accountHandler(accountUC)
```

Каждый слой получает то, что ему нужно, через конструктор. Никто не создаёт свои зависимости сам.

**Аналогия:** Директор нанимает повара и даёт ему ключи от холодильника. Повар не сам добывает ключи — ему дают при устройстве на работу.

### Роутинг (кто обрабатывает какой URL)

```go
http.HandleFunc("/api/accounts", accountHandler.HandleList)
http.HandleFunc("/api/accounts/", func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    if strings.Contains(path, "/transactions") {
        transactionHandler.Handle(w, r)  // /api/accounts/5/transactions
    } else {
        accountHandler.HandleByID(w, r)  // /api/accounts/5
    }
})
```

Go стандартная библиотека `net/http` не умеет красивый роутинг как Express.js, поэтому мы вручную проверяем: если в URL есть `/transactions` — отдаём транзакционному хендлеру, иначе — хендлеру счетов.

---

## Шаг 7: Фронтенд (Vue.js)

### `AccountsView.vue` — главная страница

#### Загрузка данных

```typescript
function loadAccounts() {
  fetch('/api/accounts')           // Делаем HTTP GET запрос
    .then((response) => response.json())  // Парсим JSON из ответа
    .then((data) => (accounts.value = data))  // Сохраняем в реактивную переменную
}
```

`fetch` — это встроенная функция браузера для HTTP-запросов. Работает через промисы (`.then`):

```
fetch(url)  →  получили ответ  →  распарсили JSON  →  сохранили в переменную
```

#### Создание счёта

```typescript
async function addAccount() {
  const response = await fetch('/api/accounts', {
    method: 'POST',                              // Метод POST = создать
    headers: { 'Content-Type': 'application/json' },  // Говорим серверу: шлём JSON
    body: JSON.stringify({                        // Тело запроса
      currency: code,
      comment: newComment.value.trim(),
    }),
  })
  if (response.ok) {     // Если сервер ответил 200-299
    loadAccounts()        // Перезагружаем список
  }
}
```

**`async/await` vs `.then`** — это два способа написать одно и то же:

```typescript
// Способ 1: .then (цепочка)
fetch(url).then(r => r.json()).then(data => ...)

// Способ 2: async/await (читается как обычный код)
const response = await fetch(url)
const data = await response.json()
```

#### Итоги по валютам

```typescript
const currencyTotals = computed(() => {
  const totals: Record<string, number> = {}
  accounts.value.forEach((a) => {
    totals[a.currency] = (totals[a.currency] ?? 0) + a.balance
  })
  return totals
})
```

`computed` — это вычисляемое свойство Vue. Автоматически пересчитывается, когда `accounts` меняется.

Пример: если есть 2 счёта в USD (баланс 100 и 200) и 1 в EUR (баланс 50):
```
{ "USD": 300, "EUR": 50 }
```

`??` — оператор "если null или undefined, то используй правое значение". `totals["USD"] ?? 0` → если USD ещё не в объекте, начинаем с 0.

### `AccountDetailView.vue` — страница одного счёта

#### Пополнение и списание

```typescript
async function deposit() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return
  await createTransaction(amount)      // +100 → пополнение
}

async function withdraw() {
  const amount = parseFloat(txAmount.value)
  if (!amount || amount <= 0) return
  await createTransaction(-amount)     // -100 → списание
}
```

Пользователь вводит `100`. Нажимает "Пополнить" → отправляется `+100`. Нажимает "Списать" → отправляется `-100`.

### Роутер (`router/index.ts`)

```typescript
{
  path: '/accounts',
  name: 'accounts',
  component: () => import('../views/AccountsView.vue'),
},
{
  path: '/accounts/:id',
  name: 'account-detail',
  component: () => import('../views/AccountDetailView.vue'),
},
```

`:id` — динамический параметр. `/accounts/5` → `id = 5`. Во Vue-компоненте получаем через `useRoute().params.id`.

`() => import(...)` — ленивая загрузка. Компонент загрузится только когда пользователь перейдёт на эту страницу.

---

## Полная карта API-эндпоинтов

| Метод | URL | Что делает | Пример тела запроса |
|-------|-----|-----------|-------------------|
| `GET` | `/api/accounts` | Все счета с балансами | — |
| `POST` | `/api/accounts` | Создать счёт | `{"currency":"USD","comment":"Копилка"}` |
| `GET` | `/api/accounts/1` | Один счёт | — |
| `DELETE` | `/api/accounts/1` | Удалить счёт + все операции | — |
| `GET` | `/api/accounts/1/transactions` | История операций | — |
| `POST` | `/api/accounts/1/transactions` | Новая операция | `{"amount":100,"comment":"Зарплата"}` |

---

## Путь запроса от клика до базы данных

Пример: пользователь нажимает "Пополнить" на 500 рублей на счёт id=3.

```
1. Браузер (Vue)
   → fetch('/api/accounts/3/transactions', { method: 'POST', body: '{"amount":500,"comment":"Зарплата"}' })

2. Handler (transaction.go)
   → Парсит URL → account_id = 3
   → Проверяет: счёт 3 существует? Да
   → Парсит JSON → amount = 500, comment = "Зарплата"

3. UseCase (transaction.go)
   → Вызывает repo.Create(transaction)

4. Repository (transaction.go)
   → INSERT INTO transactions (account_id, amount, comment) VALUES (3, 500, 'Зарплата')

5. PostgreSQL
   → Сохраняет строку, генерирует id=7, created_at='2026-02-19 12:00:00'

6. Обратный путь:
   Repository → UseCase → Handler → JSON ответ → Браузер обновляет страницу
```

---

## Что удалили и почему

| Файл | Почему удалили |
|------|---------------|
| `entity/balance.go` | Заменён на `account.go` + `transaction.go` |
| `repository/postgres/balance.go` | Заменён на `account.go` + `transaction.go` |
| `usecase/balance.go` | Заменён на `account.go` + `transaction.go` |
| `handler/balance.go` | Заменён на `account.go` + `transaction.go` |
| `views/CalculatorApiView.vue` | Заменён на `AccountsView.vue` + `AccountDetailView.vue` |

---

## Ключевые концепции — шпаргалка

### Clean Architecture (Чистая архитектура)

```
Entity      → ЧТО это (структуры данных)
Repository  → ГДЕ хранить (SQL-запросы)
UseCase     → ЧТО ДЕЛАТЬ (бизнес-правила)
Handler     → КАК ОБЩАТЬСЯ с внешним миром (HTTP)
```

Зависимости идут внутрь: Handler → UseCase → Repository → Entity. Никогда наоборот.

### Dependency Injection (Внедрение зависимостей)

Вместо:
```go
// Плохо: повар сам достаёт ключи
func NewChef() { keys := FindKeys() }
```

Делаем:
```go
// Хорошо: директор даёт повару ключи
func NewChef(keys Keys) { ... }
```

### COALESCE

```sql
COALESCE(значение, запасное_значение)
-- Если значение = NULL → вернуть запасное
-- COALESCE(NULL, 0) → 0
-- COALESCE(500, 0)  → 500
```

### ON DELETE CASCADE

```
Счёт удалён → все его транзакции удаляются автоматически
```

Как удаление папки — файлы внутри удалятся вместе с ней.

### Vue `ref` и `computed`

```typescript
const count = ref(0)        // Реактивная переменная. Меняешь → экран обновляется
const double = computed(() => count.value * 2)  // Вычисляется автоматически из других ref
```

### HTTP-методы

```
GET    = Дай мне данные        (прочитать)
POST   = Создай новую штуку    (создать)
PUT    = Обнови существующую   (изменить)
DELETE = Удали                 (удалить)
```

---

## Краткая сводка (для конспекта)

1. **Вместо одной таблицы `balances`** сделали две: `accounts` (счета) и `transactions` (операции)
2. **Баланс не хранится** в базе — вычисляется как `SUM(транзакций)` через SQL-подзапрос
3. **Каскадное удаление** — удалил счёт → транзакции удалились сами (`ON DELETE CASCADE`)
4. **4 слоя** кода: Entity → Repository → UseCase → Handler (снизу вверх)
5. **Dependency Injection** — каждый слой получает зависимости через конструктор в `main.go`
6. **Интерфейсы** в UseCase позволяют подменять реализацию (для тестов или другой БД)
7. **API**: 6 эндпоинтов — CRUD для счетов + создание/чтение транзакций
8. **Фронтенд**: 2 страницы — список счетов (`AccountsView`) и детали счёта (`AccountDetailView`)
9. **Роутер Vue**: `/accounts` → список, `/accounts/:id` → детали
10. **`fetch`** в Vue для общения с Go-бэкендом через JSON
