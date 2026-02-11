# Гайд: Переделываем калькулятор на автоматические курсы валют

## Оглавление

1. [Что у нас есть сейчас](#1-что-у-нас-есть-сейчас)
2. [Что нужно получить в итоге](#2-что-нужно-получить-в-итоге)
3. [Общая картина — как всё будет работать](#3-общая-картина--как-всё-будет-работать)
4. [Шаг 1 — Выбираем сторонний API для курсов](#шаг-1--выбираем-сторонний-api-для-курсов)
5. [Шаг 2 — Меняем базу данных (SQLite)](#шаг-2--меняем-базу-данных-sqlite)
6. [Шаг 3 — Меняем Go-сервер (Main.go)](#шаг-3--меняем-go-сервер-maingo)
7. [Шаг 4 — Меняем фронтенд (CalculatorApiView.vue)](#шаг-4--меняем-фронтенд-calculatorapiviewvue)
8. [Как тестировать](#8-как-тестировать)
9. [Словарик терминов](#9-словарик-терминов)

---

## 1. Что у нас есть сейчас

Представь, что наше приложение — это записная книжка про деньги.

**Сейчас работает так:**

```
Пользователь вводит:  "EUR", сумма: 100, курс: 1.1
                      "GBP", сумма: 50,  курс: 1.27
```

То есть пользователь **сам руками вводит курс** каждой валюты. Это неудобно, потому что:
- Курсы меняются каждую секунду
- Пользователь может ввести неправильный курс
- Нужно самому гуглить актуальные курсы

**Как устроен текущий код:**

```
ФРОНТЕНД (Vue)                    СЕРВЕР (Go)                БАЗА ДАННЫХ (SQLite)
┌──────────────────┐          ┌──────────────────┐          ┌──────────────────┐
│ Пользователь     │          │                  │          │ Таблица balances │
│ вводит:          │  ──────> │  Main.go          │  ──────> │                  │
│  - валюту (EUR)  │  запрос  │  принимает данные │  пишет   │  id | currency   │
│  - сумму (100)   │          │  и отдаёт их     │  в базу  │     | amount     │
│  - курс (1.1)    │  <────── │                  │  <────── │     | rate       │
│                  │  ответ   │                  │  читает  │                  │
└──────────────────┘          └──────────────────┘          └──────────────────┘
```

**Текущая таблица в базе данных:**
```
id  | currency | amount | rate
----|----------|--------|------
1   | EUR      | 100    | 1.1
2   | GBP      | 50     | 1.27
```

Тут `rate` (курс) хранится ВМЕСТЕ с балансом. Пользователь вводит его руками.

---

## 2. Что нужно получить в итоге

**Новая логика:**
- Пользователь вводит **только валюту и сумму** (например: "EUR", 100)
- **Курс вводить НЕ НУЖНО** — сервер сам получает курсы из интернета
- Сервер каждые 30 секунд ходит на сторонний сайт за свежими курсами
- Курсы хранятся отдельно в своей таблице в базе
- Если для какой-то валюты курса нет — на экране показывается предупреждение

---

## 3. Общая картина — как всё будет работать

```
СТОРОННИЙ API                  НАШ СЕРВЕР (Go)               БАЗА ДАННЫХ (SQLite)
(exchangerate-api.com)
┌──────────────────┐          ┌──────────────────┐          ┌──────────────────────┐
│                  │          │                  │          │ Таблица balances:    │
│  Даёт актуальные │  <────── │  Каждые 30 сек   │  ──────> │   id | currency      │
│  курсы валют     │  запрос  │  Go-сервер       │  пишет   │      | amount        │
│  всего мира      │  ──────> │  спрашивает      │          │                      │
│                  │  ответ   │  курсы           │          │ Таблица rates:       │
│                  │          │                  │  ──────> │   id | currency      │
└──────────────────┘          │                  │  пишет   │      | rate          │
                              │                  │          │      | updated_at    │
   ФРОНТЕНД (Vue)            │                  │          │                      │
┌──────────────────┐          │                  │          └──────────────────────┘
│ Пользователь     │          │                  │
│ вводит:          │  ──────> │  Два эндпоинта:  │
│  - валюту (EUR)  │          │  /api/balances   │
│  - сумму (100)   │          │  /api/rates      │
│                  │  <────── │                  │
│ Видит курс       │          │                  │
│ автоматически!   │          │                  │
└──────────────────┘          └──────────────────┘
```

**Было 1 таблица, стало 2 таблицы:**

Таблица `balances` (что есть у пользователя):
```
id  | currency | amount
----|----------|--------
1   | EUR      | 100
2   | GBP      | 50
```

Таблица `rates` (курсы из интернета):
```
id  | currency | rate_to_usd | updated_at
----|----------|-------------|--------------------
1   | EUR      | 1.08        | 2026-02-11 12:00:00
2   | GBP      | 1.26        | 2026-02-11 12:00:00
3   | JPY      | 0.0067      | 2026-02-11 12:00:00
...тут будут ВСЕ валюты мира...
```

---

## Шаг 1 — Выбираем сторонний API для курсов

### Что такое "сторонний API"?

Представь, что есть сайт, который знает все курсы валют в мире. Ты можешь отправить ему запрос (как будто заходишь на страницу), и он тебе ответит — вот тебе все курсы в формате JSON.

### Какой API использовать?

Рекомендую **ExchangeRate-API** — он бесплатный (до 1500 запросов в месяц) и простой.

**Сайт:** https://www.exchangerate-api.com/

### Как получить доступ:

1. Зайди на https://www.exchangerate-api.com/
2. Нажми "Get Free Key" (получить бесплатный ключ)
3. Введи свой email, придумай пароль
4. Тебе на почту придёт **API ключ** — это как пароль для доступа к данным

API ключ выглядит примерно так: `a1b2c3d4e5f6g7h8i9j0`

### Как проверить, что API работает:

Открой в браузере эту ссылку (подставь свой ключ вместо `ТВОЙ_КЛЮЧ`):

```
https://v6.exchangerate-api.com/v6/ТВОЙ_КЛЮЧ/latest/USD
```

Ты увидишь что-то вроде:

```json
{
  "result": "success",
  "base_code": "USD",
  "conversion_rates": {
    "USD": 1,
    "EUR": 0.92,
    "GBP": 0.79,
    "JPY": 149.5,
    "RUB": 92.5
    ... и ещё ~160 валют
  }
}
```

Это и есть те курсы, которые наш сервер будет забирать каждые 30 секунд.

> **Важно:** `conversion_rates` — это курс относительно USD. То есть `"EUR": 0.92` значит
> "за 1 доллар дают 0.92 евро". Нам в коде нужно будет **перевернуть** это число,
> чтобы получить "сколько долларов стоит 1 евро" (1 / 0.92 = 1.087).
> Либо можно хранить как есть и просто адаптировать формулу расчёта.

---

## Шаг 2 — Меняем базу данных (SQLite)

### Что нужно сделать:

1. **Убрать** колонку `rate` из таблицы `balances`
2. **Создать** новую таблицу `rates`

### 2.1. Новая таблица `balances` (без rate)

Раньше было:
```sql
CREATE TABLE balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    current TEXT NOT NULL,
    amount REAL NOT NULL,
    rate REAL NOT NULL          -- ЭТО УДАЛЯЕМ
);
```

Станет:
```sql
CREATE TABLE balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    currency TEXT NOT NULL,
    amount REAL NOT NULL
);
```

> **Обрати внимание:** в старой таблице колонка называлась `current`, а должна `currency`.
> Это, похоже, была опечатка в оригинале. Исправь на `currency`.

### 2.2. Новая таблица `rates`

```sql
CREATE TABLE IF NOT EXISTS rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    currency TEXT NOT NULL UNIQUE,
    rate_to_usd REAL NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Объяснение каждой строки:**
- `id` — порядковый номер записи (создаётся автоматически)
- `currency TEXT NOT NULL UNIQUE` — код валюты (EUR, GBP...). `UNIQUE` значит что нельзя добавить две записи с одной валютой — при обновлении старая перезапишется
- `rate_to_usd REAL NOT NULL` — курс к доллару (дробное число). Например, для EUR это будет 1.087 (1 евро = 1.087 доллара)
- `updated_at` — когда этот курс был обновлён (дата и время)

### 2.3. Как применить изменения

**Самый простой способ** — удалить старую базу и дать серверу создать новую:

1. Останови Go-сервер (Ctrl+C в терминале)
2. Удали файл `balances.db` в корне проекта
3. Когда сервер запустится заново — он создаст новые таблицы сам (мы это пропишем в коде)

> **Внимание:** Все старые данные (балансы) будут удалены. Это нормально для учебного
> проекта. В реальном проекте делали бы "миграцию" — но это пока сложная тема.

---

## Шаг 3 — Меняем Go-сервер (Main.go)

Это самый большой шаг. Разобьём его на маленькие части.

### 3.1. Добавляем новые импорты

В начале файла `Main.go` нужно добавить пакеты для работы с HTTP-запросами и таймером.

Было:
```go
import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)
```

Станет:
```go
import (
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"
)
```

**Что добавилось:**
- `"io"` — для чтения ответа от стороннего API (чтение "тела" ответа)
- `"time"` — для таймера (каждые 30 секунд)

### 3.2. Добавляем константу с API ключом

Прямо после импортов, добавь:

```go
const exchangeRateAPIKey = "СЮДА_ВСТАВЬ_СВОЙ_КЛЮЧ"
const exchangeRateAPIURL = "https://v6.exchangerate-api.com/v6/"
```

> **Важно для безопасности:** В реальных проектах ключи НИКОГДА не хранят прямо в коде.
> Их кладут в файл `.env` или переменные окружения. Но для учёбы — сойдёт.

### 3.3. Меняем структуры данных

Раньше была одна структура:
```go
type Balance struct {
    ID       int     `json:"id"`
    Currency string  `json:"currency"`
    Amount   float64 `json:"amount"`
    Rate     float64 `json:"rate"`
}
```

Теперь нужны **две** структуры:

```go
// Balance — сколько денег у пользователя в какой валюте
type Balance struct {
    ID       int     `json:"id"`
    Currency string  `json:"currency"`
    Amount   float64 `json:"amount"`
}

// Rate — курс валюты (приходит с сервера автоматически)
type Rate struct {
    ID        int     `json:"id"`
    Currency  string  `json:"currency"`
    RateToUSD float64 `json:"rate_to_usd"`
    UpdatedAt string  `json:"updated_at"`
}
```

А также структура для разбора ответа от стороннего API:

```go
// ExchangeRateResponse — формат ответа от exchangerate-api.com
type ExchangeRateResponse struct {
    Result          string             `json:"result"`
    ConversionRates map[string]float64 `json:"conversion_rates"`
}
```

**Что тут происходит:**
- `ExchangeRateResponse` описывает JSON, который придёт от стороннего API
- `ConversionRates` — это `map[string]float64`, то есть словарь где ключ — название валюты (string), а значение — курс (число)
- Например: `{"EUR": 0.92, "GBP": 0.79}` превратится в Go-шный map

### 3.4. Меняем функцию `createTable()`

Теперь нужно создавать **две** таблицы:

```go
func createTable() {
    // Таблица балансов пользователя (БЕЗ курса!)
    queryBalances := `
    CREATE TABLE IF NOT EXISTS balances (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        currency TEXT NOT NULL,
        amount REAL NOT NULL
    );`

    _, err := db.Exec(queryBalances)
    if err != nil {
        log.Fatal("Ошибка создания таблицы balances:", err)
    }

    // Таблица курсов валют (заполняется автоматически)
    queryRates := `
    CREATE TABLE IF NOT EXISTS rates (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        currency TEXT NOT NULL UNIQUE,
        rate_to_usd REAL NOT NULL,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

    _, err = db.Exec(queryRates)
    if err != nil {
        log.Fatal("Ошибка создания таблицы rates:", err)
    }
}
```

### 3.5. Добавляем функцию получения курсов из стороннего API

Это самая новая и важная часть. Эта функция:
1. Отправляет запрос на exchangerate-api.com
2. Получает ответ с курсами
3. Сохраняет каждый курс в нашу таблицу `rates`

```go
// fetchAndSaveRates — ходит в сторонний API и сохраняет курсы в базу
func fetchAndSaveRates() {
    // 1. Составляем URL для запроса
    url := exchangeRateAPIURL + exchangeRateAPIKey + "/latest/USD"

    // 2. Отправляем GET-запрос (как будто открываем страницу в браузере)
    resp, err := http.Get(url)
    if err != nil {
        log.Println("Ошибка запроса к API курсов:", err)
        return  // не падаем, просто пишем ошибку в лог
    }
    defer resp.Body.Close()  // закрываем соединение когда закончим

    // 3. Читаем тело ответа (это текст в формате JSON)
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Println("Ошибка чтения ответа API:", err)
        return
    }

    // 4. Превращаем JSON-текст в Go-структуру
    var ratesResponse ExchangeRateResponse
    err = json.Unmarshal(body, &ratesResponse)
    if err != nil {
        log.Println("Ошибка разбора JSON от API:", err)
        return
    }

    // 5. Проверяем что API ответил успешно
    if ratesResponse.Result != "success" {
        log.Println("API вернул ошибку, result:", ratesResponse.Result)
        return
    }

    // 6. Сохраняем каждый курс в базу данных
    for currency, rate := range ratesResponse.ConversionRates {
        // rate — это "сколько этой валюты дают за 1 USD"
        // Нам нужно "сколько USD стоит 1 единица этой валюты"
        // Поэтому делаем 1 / rate
        var rateToUSD float64
        if rate > 0 {
            rateToUSD = 1.0 / rate
        }

        // INSERT OR REPLACE — если валюта уже есть, перезаписываем курс
        // Если валюты нет — создаём новую запись
        _, err := db.Exec(`
            INSERT INTO rates (currency, rate_to_usd, updated_at)
            VALUES (?, ?, CURRENT_TIMESTAMP)
            ON CONFLICT(currency)
            DO UPDATE SET rate_to_usd = ?, updated_at = CURRENT_TIMESTAMP
        `, currency, rateToUSD, rateToUSD)

        if err != nil {
            log.Println("Ошибка сохранения курса для", currency, ":", err)
        }
    }

    log.Println("Курсы валют обновлены успешно!")
}
```

**Разбор по строчкам:**

- `http.Get(url)` — это как когда ты в браузере открываешь ссылку. Go делает то же самое, только без браузера
- `defer resp.Body.Close()` — "когда функция закончится, закрой соединение". `defer` = "отложи это на потом"
- `io.ReadAll(resp.Body)` — прочитай весь текст ответа и положи в переменную `body`
- `json.Unmarshal(body, &ratesResponse)` — превращает текст JSON в Go-структуру. Это как `JSON.parse()` в JavaScript
- `for currency, rate := range ratesResponse.ConversionRates` — проходит по каждой валюте в словаре. `currency` = ключ (например "EUR"), `rate` = значение (например 0.92)
- `ON CONFLICT(currency) DO UPDATE SET ...` — это SQL-магия: "если такая валюта уже есть — обнови её, а не создавай дубликат"

### 3.6. Добавляем таймер (каждые 30 секунд)

В функцию `main()` нужно добавить запуск периодического обновления курсов.

**Что такое горутина (goroutine)?**

Представь, что ты можешь одновременно и готовить еду, и смотреть телевизор. В Go горутина — это как "параллельная задача". Сервер будет одновременно:
- Отвечать на запросы пользователей
- Каждые 30 секунд обновлять курсы

```go
// startRateUpdater запускает фоновое обновление курсов каждые 30 секунд
func startRateUpdater() {
    // Сразу обновляем курсы при старте сервера (не ждём 30 секунд)
    fetchAndSaveRates()

    // Создаём "тикер" — он будет "тикать" каждые 30 секунд
    ticker := time.NewTicker(30 * time.Second)

    // Запускаем горутину (параллельную задачу)
    go func() {
        for range ticker.C {
            // Каждые 30 секунд вызываем функцию обновления курсов
            fetchAndSaveRates()
        }
    }()
}
```

### 3.7. Добавляем новый эндпоинт `/api/rates`

Фронтенду нужен способ получить курсы. Создаём новый обработчик:

```go
// getRates — отдаёт все курсы из базы данных
func getRates(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    rows, err := db.Query("SELECT id, currency, rate_to_usd, updated_at FROM rates")
    if err != nil {
        http.Error(w, `{"error": "Ошибка получения курсов"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    rates := []Rate{}
    for rows.Next() {
        var rate Rate
        if err := rows.Scan(&rate.ID, &rate.Currency, &rate.RateToUSD, &rate.UpdatedAt); err != nil {
            http.Error(w, `{"error": "Ошибка чтения курсов"}`, http.StatusInternalServerError)
            return
        }
        rates = append(rates, rate)
    }

    json.NewEncoder(w).Encode(rates)
}
```

### 3.8. Меняем функции работы с балансами

Поскольку в `balances` больше нет `rate`, нужно обновить функции.

**`getBalances` — убираем rate из запроса:**

```go
func getBalances(w http.ResponseWriter, _ *http.Request) {
    rows, err := db.Query("SELECT id, currency, amount FROM balances")
    if err != nil {
        http.Error(w, `{"error": "Ошибка получения балансов"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    balances := []Balance{}
    for rows.Next() {
        var balance Balance
        if err := rows.Scan(&balance.ID, &balance.Currency, &balance.Amount); err != nil {
            http.Error(w, `{"error": "Ошибка чтения данных"}`, http.StatusInternalServerError)
            return
        }
        balances = append(balances, balance)
    }

    json.NewEncoder(w).Encode(balances)
}
```

**`createBalance` — убираем rate:**

```go
func createBalance(w http.ResponseWriter, r *http.Request) {
    var balance Balance
    if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
        http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
        return
    }

    result, err := db.Exec(
        "INSERT INTO balances (currency, amount) VALUES (?, ?)",
        balance.Currency, balance.Amount,
    )
    if err != nil {
        http.Error(w, `{"error": "Ошибка создания баланса"}`, http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    balance.ID = int(id)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(balance)
}
```

**`getBalance` — убираем rate:**

```go
func getBalance(w http.ResponseWriter, r *http.Request, id int) {
    var balance Balance
    err := db.QueryRow("SELECT id, currency, amount FROM balances WHERE id = ?", id).
        Scan(&balance.ID, &balance.Currency, &balance.Amount)

    if err == sql.ErrNoRows {
        http.Error(w, `{"error": "Баланс не найден"}`, http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, `{"error": "Ошибка получения баланса"}`, http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(balance)
}
```

**`updateBalance` — теперь обновляем только amount:**

```go
func updateBalance(w http.ResponseWriter, r *http.Request, id int) {
    var balance Balance
    if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
        http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
        return
    }

    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE id = ?)", id).Scan(&exists)
    if err != nil || !exists {
        http.Error(w, `{"error": "Баланс не найден"}`, http.StatusNotFound)
        return
    }

    _, err = db.Exec(
        "UPDATE balances SET currency = ?, amount = ? WHERE id = ?",
        balance.Currency, balance.Amount, id,
    )
    if err != nil {
        http.Error(w, `{"error": "Ошибка обновления баланса"}`, http.StatusInternalServerError)
        return
    }

    balance.ID = id
    json.NewEncoder(w).Encode(balance)
}
```

**`deleteBalance` — остаётся почти без изменений** (только сообщения об ошибках, если хочешь).

### 3.9. Меняем функцию `main()`

```go
func main() {
    var err error
    db, err = sql.Open("sqlite3", "./balances.db")
    if err != nil {
        log.Fatal("Ошибка подключения к БД", err)
    }
    defer db.Close()

    // Создание таблиц (balances + rates)
    createTable()

    // Запускаем фоновое обновление курсов
    startRateUpdater()

    // Маршруты для балансов (как раньше)
    http.HandleFunc("/api/balances", handleBalancesList)
    http.HandleFunc("/api/balances/", handleBalanceById)

    // НОВЫЙ маршрут для курсов
    http.HandleFunc("/api/rates", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            getRates(w, r)
        } else {
            http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
        }
    })

    fmt.Println("Сервер запущен на http://localhost:8080")
    fmt.Println("Endpoints:")
    fmt.Println("  GET    /api/balances     - список балансов")
    fmt.Println("  POST   /api/balances     - создать баланс")
    fmt.Println("  GET    /api/balances/{id} - получить баланс")
    fmt.Println("  PUT    /api/balances/{id} - обновить баланс")
    fmt.Println("  DELETE /api/balances/{id} - удалить баланс")
    fmt.Println("  GET    /api/rates         - список курсов (НОВОЕ!)")

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## Шаг 4 — Меняем фронтенд (CalculatorApiView.vue)

### 4.1. Что меняется на фронте:

1. Убираем поле ввода "Rate" из формы
2. Добавляем загрузку курсов с сервера (`/api/rates`)
3. Курс отображается автоматически рядом с каждой валютой
4. Если курса нет — показываем предупреждение
5. При добавлении баланса отправляем только `currency` и `amount` (без `rate`)

### 4.2. Меняем `<script setup>`

```vue
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import LightLayout from '@/layouts/LightLayout.vue'

// ---- ТИПЫ ----

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

// ---- СОСТОЯНИЕ (reactive data) ----

const currencies = ref<Balance[]>([])   // балансы пользователя
const rates = ref<Rate[]>([])           // курсы валют с сервера
const newCurrency = ref('')             // поле ввода: код валюты
const newAmount = ref('')               // поле ввода: сумма
// newRate больше НЕ НУЖЕН — курс приходит автоматически!

// ---- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ----

// Найти курс для конкретной валюты
// Возвращает число (курс) или null если курса нет
function getRateForCurrency(currencyCode: string): number | null {
  const found = rates.value.find(r => r.currency === currencyCode)
  return found ? found.rate_to_usd : null
}

// Проверяем: есть ли валюты, для которых нет курса?
const hasMissingRates = computed((): boolean => {
  return currencies.value.some(c => getRateForCurrency(c.currency) === null)
})

// Считаем общую сумму в USD
const totalUSD = computed((): number => {
  let total = 0
  currencies.value.forEach(c => {
    const rate = getRateForCurrency(c.currency)
    if (c.amount && rate) {
      total += c.amount * rate
    }
  })
  return total
})

// Можно ли посчитать полную сумму? (все ли курсы есть)
const canCalculateTotal = computed((): boolean => {
  // Если хотя бы у одной валюты с суммой > 0 нет курса, то нельзя
  return !currencies.value.some(c => {
    return c.amount && c.amount > 0 && getRateForCurrency(c.currency) === null
  })
})

// ---- ЗАГРУЗКА ДАННЫХ ----

// Загружаем балансы пользователя
function loadBalances() {
  fetch('/api/balances')
    .then(response => response.json())
    .then(data => currencies.value = data)
}

// Загружаем курсы с сервера
function loadRates() {
  fetch('/api/rates')
    .then(response => response.json())
    .then(data => rates.value = data)
}

// При загрузке страницы — грузим и балансы, и курсы
onMounted(() => {
  loadBalances()
  loadRates()

  // Обновляем курсы каждые 30 секунд (чтобы на экране были свежие)
  setInterval(loadRates, 30000)
})

// ---- ДЕЙСТВИЯ ПОЛЬЗОВАТЕЛЯ ----

async function addCurrency() {
  const code = newCurrency.value.trim().toUpperCase()
  if (!code) return

  const response = await fetch('/api/balances', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      currency: code,
      amount: parseFloat(newAmount.value) || 0
      // rate больше НЕ отправляем!
    })
  })

  if (response.ok) {
    newCurrency.value = ''
    newAmount.value = ''
    loadBalances()
  }
}

async function removeBalance(balance: Balance) {
  const response = await fetch(`/api/balances/${balance.id}`, {
    method: 'DELETE'
  })
  if (response.ok) {
    loadBalances()
  }
}

async function updateBalance(balance: Balance) {
  await fetch(`/api/balances/${balance.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(balance)
    // rate не отправляем — сервер сам знает курсы
  })
}
</script>
```

### 4.3. Меняем `<template>`

```vue
<template>
  <LightLayout>
    <div class="calculator-container">
      <div class="calculator-card">
        <h1 class="calculator-title">Currency Calculator</h1>

        <!-- Предупреждение если нет курсов или не хватает -->
        <div v-if="rates.length === 0" class="warning-banner">
          Курсы валют не загружены. Расчёт невозможен.
        </div>
        <div v-else-if="hasMissingRates" class="warning-banner">
          Для некоторых валют нет курса. Общий расчёт может быть неточным.
        </div>

        <!-- Форма добавления (без поля Rate!) -->
        <form @submit.prevent="addCurrency" class="add-form">
          <input
            v-model="newCurrency"
            placeholder="Currency (USD, EUR...)"
            class="input-field input-small"
          />
          <input
            v-model="newAmount"
            placeholder="Amount"
            class="input-field input-small"
          />
          <!-- Поля Rate больше нет! -->
          <button type="submit" class="btn btn-primary">+ Add</button>
        </form>

        <div v-if="currencies.length === 0" class="empty-state">
          No currencies yet. Add one above!
        </div>

        <!-- Список валют -->
        <div class="currencies-list">
          <div
            v-for="balance in currencies"
            :key="balance.id"
            class="currency-item"
          >
            <span class="currency-code">{{ balance.currency }}</span>

            <input
              v-model="balance.amount"
              @change="updateBalance(balance)"
              placeholder="Amount"
              type="number"
              class="input-field input-small"
            />

            <!-- Показываем курс автоматически (или предупреждение) -->
            <span
              v-if="getRateForCurrency(balance.currency) !== null"
              class="rate-display"
            >
              1 {{ balance.currency }} = {{ getRateForCurrency(balance.currency)!.toFixed(4) }} USD
            </span>
            <span v-else class="rate-missing">
              нет курса
            </span>

            <button @click="removeBalance(balance)" class="btn btn-danger">
              ✕
            </button>
          </div>
        </div>

        <!-- Итого -->
        <div v-if="currencies.length > 0" class="total-section">
          <h2 class="total-title">Total Conversion</h2>

          <div v-if="!canCalculateTotal" class="warning-banner warning-small">
            Не все курсы доступны — итог может быть неполным
          </div>

          <ul class="total-list">
            <li
              v-if="!currencies.some(c => c.currency === 'USD') && totalUSD"
              class="total-item"
            >
              <span class="total-currency">USD</span>
              <span class="total-value">{{ totalUSD.toFixed(2) }}</span>
            </li>
            <template v-for="currency in currencies" :key="currency.currency">
              <li
                v-if="getRateForCurrency(currency.currency)"
                class="total-item"
              >
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
```

### 4.4. Добавляем стили для новых элементов

В секцию `<style scoped>` **добавь** (не удаляй старые стили, просто допиши):

```css
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
```

---

## 8. Как тестировать

### Порядок запуска:

1. **Удали старую базу** (файл `balances.db` в корне проекта)

2. **Запусти Go-сервер:**
   ```bash
   go run Main.go
   ```
   В консоли должно появиться:
   ```
   Курсы валют обновлены успешно!
   Сервер запущен на http://localhost:8080
   ```

3. **Запусти фронтенд** (в другом терминале):
   ```bash
   npm run dev
   ```

4. **Открой в браузере** http://localhost:5173/calc-api

### Что проверить:

1. **Курсы загружаются?**
   Открой в браузере: http://localhost:8080/api/rates
   Должен появиться JSON с кучей валют и курсов.

2. **Добавление баланса работает?**
   Введи "EUR" и "100", нажми "+ Add".
   Рядом с EUR должен автоматически появиться курс (зелёным).

3. **Предупреждение работает?**
   Введи валюту, которой не существует (например "XXX").
   Должна появиться красная надпись "нет курса" и жёлтое предупреждение сверху.

4. **Итоги считаются?**
   Добавь несколько реальных валют (EUR, GBP, JPY).
   В секции "Total Conversion" должны быть пересчёты.

5. **Курсы обновляются?**
   Подожди 30 секунд, проверь в консоли Go-сервера — должно снова появиться
   "Курсы валют обновлены успешно!".

---

## 9. Словарик терминов

| Термин | Что это значит |
|--------|---------------|
| **API** | "Дверь" в программу, через которую другие программы могут общаться с ней. Как окошко в банке — подаёшь запрос, получаешь ответ |
| **Эндпоинт (endpoint)** | Конкретный адрес API. Например `/api/rates` — это один эндпоинт, `/api/balances` — другой |
| **JSON** | Формат текста для обмена данными. Выглядит как `{"ключ": "значение"}` |
| **HTTP GET** | Запрос "дай мне данные" (как открыть страницу) |
| **HTTP POST** | Запрос "сохрани новые данные" |
| **HTTP PUT** | Запрос "обнови существующие данные" |
| **HTTP DELETE** | Запрос "удали данные" |
| **SQLite** | Лёгкая база данных, которая хранится в одном файле |
| **Горутина (goroutine)** | Параллельная задача в Go. Как второй работник, который делает своё дело одновременно |
| **Тикер (ticker)** | Таймер, который "тикает" через заданный интервал |
| **map** | Словарь/справочник. Хранит пары "ключ: значение" |
| **UNIQUE** | Ограничение в базе — нельзя добавить два одинаковых значения |
| **ON CONFLICT** | SQL-инструкция: "если запись уже существует — обнови, а не создавай дубликат" |
| **defer** | Ключевое слово Go: "выполни это позже, когда функция закончится" |
| **ref** | Реактивная переменная во Vue. Когда она меняется — экран обновляется сам |
| **computed** | Вычисляемое значение во Vue. Пересчитывается автоматически когда меняются данные, от которых зависит |
| **setInterval** | JavaScript-таймер: "повторяй эту функцию каждые N миллисекунд" |
| **rate_to_usd** | Сколько долларов стоит 1 единица валюты. Например, rate_to_usd для EUR = 1.087 значит 1 евро = 1.087 доллара |
