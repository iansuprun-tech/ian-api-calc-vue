# Что изменилось: Авторизация + Редизайн UI

## Оглавление

1. [Общая картина — было / стало](#1-общая-картина--было--стало)
2. [База данных — новые таблицы](#2-база-данных--новые-таблицы)
3. [Backend — авторизация (как работает)](#3-backend--авторизация)
4. [Backend — привязка данных к пользователю](#4-backend--привязка-данных-к-пользователю)
5. [Frontend — авторизация](#5-frontend--авторизация)
6. [Frontend — редизайн](#6-frontend--редизайн)
7. [Курсы валют — новый интервал](#7-курсы-валют)
8. [Примеры запросов (curl)](#8-примеры-запросов-curl)
9. [Карта файлов — что добавлено, изменено, удалено](#9-карта-файлов)

---

## 1. Общая картина — было / стало

### Было

```
Любой человек заходит на сайт
        │
        ▼
  GET /api/accounts
        │
        ▼
  Видит ВСЕ счета в системе
  (общие для всех, нет понятия "мой")
```

- Нет регистрации, нет входа
- Один человек создал счёт — другой его видит и может удалить
- На главной странице — логотип Vue, "You did it!", ссылки Home / About
- Курсы валют обновляются каждые 5 минут (слишком часто для бесплатного API)

### Стало

```
Пользователь А                          Пользователь Б
     │                                        │
  Регистрация                              Регистрация
  (email + пароль)                         (email + пароль)
     │                                        │
  Вход → получает                          Вход → получает
  JWT-токен                                JWT-токен
     │                                        │
     ▼                                        ▼
  GET /api/accounts                      GET /api/accounts
  + токен А                              + токен Б
     │                                        │
     ▼                                        ▼
  Видит только                           Видит только
  СВОИ счета                             СВОИ счета
```

- Регистрация и вход с JWT-токенами
- Каждый пользователь видит **только свои** счета
- Приложение выглядит как реальный финансовый сервис (topbar "FinTrack", карточки)
- Курсы обновляются **раз в час**

---

## 2. База данных — новые таблицы

### Миграция 000005 — таблица `users`

```sql
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,        -- уникальный номер (1, 2, 3...)
    email         TEXT NOT NULL UNIQUE,       -- email, не может повторяться
    password_hash TEXT NOT NULL,              -- хеш пароля (НЕ сам пароль!)
    created_at    TIMESTAMP DEFAULT NOW()    -- когда зарегистрировался
);
```

**Зачем `password_hash`, а не `password`?**

Пароли **никогда** не хранят в открытом виде. Если кто-то получит доступ к базе, он увидит:

```
email            │ password_hash
─────────────────┼──────────────────────────────────────
ivan@mail.com    │ $2a$10$xK3f8j9Qm2kp7Hd...  ← это bcrypt-хеш, 60 символов
anna@mail.com    │ $2a$10$9Qm2kp4Lz8mN3Xf...  ← из хеша нельзя восстановить пароль
```

Bcrypt — это алгоритм хеширования. Работает в одну сторону:
```
"mypassword"  ──bcrypt──►  "$2a$10$xK3f8j..."  (легко)
"$2a$10$xK3f8j..."  ──?──►  "mypassword"       (невозможно)
```

При входе мы не расшифровываем хеш, а хешируем введённый пароль и сравниваем хеши.

### Миграция 000006 — колонка `user_id` в accounts

```sql
ALTER TABLE accounts ADD COLUMN user_id INTEGER REFERENCES users(id);
```

Теперь каждый счёт «знает», кому он принадлежит:

```
БЫЛО (таблица accounts):
id │ currency │ comment
───┼──────────┼─────────────
 1 │ USD      │ Основной        ← чей? Непонятно
 2 │ EUR      │ Отпуск          ← чей? Непонятно

СТАЛО (таблица accounts):
id │ currency │ comment      │ user_id
───┼──────────┼──────────────┼────────
 1 │ USD      │ Основной     │ 1        ← принадлежит Ивану
 2 │ EUR      │ Отпуск       │ 1        ← принадлежит Ивану
 3 │ RUB      │ Зарплата     │ 2        ← принадлежит Анне
```

`REFERENCES users(id)` — это внешний ключ (foreign key). БД не позволит записать `user_id = 999`, если пользователя с id=999 нет.

---

## 3. Backend — авторизация

### Как работает регистрация — по шагам

```
Браузер                                Сервер (Go)                         БД
   │                                       │                                │
   │  POST /api/register                   │                                │
   │  {"email":"ivan@mail.com",            │                                │
   │   "password":"secret123"}             │                                │
   │ ─────────────────────────────────►    │                                │
   │                                       │                                │
   │                          1. Получает email и password                  │
   │                          2. Хеширует пароль через bcrypt:              │
   │                             "secret123" → "$2a$10$xK3..."             │
   │                          3. INSERT INTO users                         │
   │                             (email, password_hash)  ─────────────►    │
   │                                       │                                │
   │                                       │   ◄──── id=1, created_at      │
   │                                       │                                │
   │  ◄── 201 Created ────────────────     │                                │
   │  {"id":1, "email":"ivan@mail.com"}    │                                │
```

**Файл:** `internal/usecase/auth.go` → метод `Register`

```go
func (uc *AuthUseCase) Register(email, password string) (entity.User, error) {
    // Хешируем пароль (bcrypt добавляет «соль» — случайные байты)
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    // hash = "$2a$10$xK3f8j9Qm2kp7Hd..."

    // Сохраняем в БД (email + хеш, НЕ пароль)
    return uc.repo.Create(email, string(hash))
}
```

### Как работает вход — по шагам

```
Браузер                                Сервер                              БД
   │                                       │                                │
   │  POST /api/login                      │                                │
   │  {"email":"ivan@mail.com",            │                                │
   │   "password":"secret123"}             │                                │
   │ ─────────────────────────────────►    │                                │
   │                                       │                                │
   │                          1. Ищет пользователя по email ──────────►    │
   │                                       │                                │
   │                                       │   ◄── user (id=1, hash=...)   │
   │                                       │                                │
   │                          2. Сравнивает: bcrypt.Compare(               │
   │                               хеш_из_базы,                            │
   │                               "secret123"                             │
   │                             ) → совпало!                              │
   │                                       │                                │
   │                          3. Создаёт JWT-токен:                        │
   │                             { user_id: 1, email: "...", exp: ... }    │
   │                             + подписывает секретным ключом            │
   │                                       │                                │
   │  ◄── 200 OK ─────────────────────     │                                │
   │  {"token": "eyJhbGci..."}             │                                │
```

### Что внутри JWT-токена?

JWT — это строка из трёх частей, разделённых точками:

```
eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6Iml2YW5AbWFpbC5jb20iLCJleHAiOjE3MDg1MjQ4MDB9.xxxxx
└──── заголовок ─────┘└───────────────────── данные (payload) ──────────────────────┘└─ подпись ─┘
```

Средняя часть (payload) — это Base64-закодированный JSON:

```json
{
  "user_id": 1,                 // ID пользователя — по нему фильтруем данные
  "email": "ivan@mail.com",
  "exp": 1740227200             // срок действия — через 72 часа
}
```

**Подпись** — гарантия, что токен не подделан. Сервер знает секретный ключ и может проверить подпись. Если кто-то изменит `user_id: 1` на `user_id: 2` — подпись не совпадёт, и сервер отклонит токен.

### JWT Middleware — как защищены маршруты

Middleware — это «фильтр», который стоит перед обработчиком запроса:

```
Запрос от браузера
        │
        ▼
  ┌─────────────────────────────────┐
  │        AuthMiddleware            │
  │                                  │
  │  1. Есть заголовок               │
  │     Authorization: Bearer XXX?   │
  │     Нет → 401 "Требуется        │
  │            авторизация"          │
  │                                  │
  │  2. Парсим JWT-токен             │
  │     Невалидный → 401             │
  │                                  │
  │  3. Извлекаем user_id из токена  │
  │     Кладём в context запроса     │
  │                                  │
  │  4. Пропускаем дальше ────────── │ ──► Handler
  └─────────────────────────────────┘
```

**Файл:** `internal/handler/middleware.go`

```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Читаем заголовок "Authorization: Bearer eyJhbGci..."
        authHeader := r.Header.Get("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 2. Парсим и проверяем JWT
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil  // проверяем подпись секретным ключом
        })

        // 3. Извлекаем user_id и кладём в context
        claims := token.Claims.(jwt.MapClaims)
        userID := int(claims["user_id"].(float64))
        ctx := context.WithValue(r.Context(), userIDKey, userID)

        // 4. Передаём дальше с обновлённым context
        next(w, r.WithContext(ctx))
    }
}
```

В handler'е достаём user_id:

```go
func (h *AccountHandler) HandleList(w http.ResponseWriter, r *http.Request) {
    userID, _ := UserIDFromContext(r.Context())
    // userID = 1 (из токена Ивана)
    // userID = 2 (из токена Анны)

    accounts, _ := h.uc.GetAll(userID)
    // → SELECT ... FROM accounts WHERE user_id = 1
}
```

### Какие маршруты защищены?

```
/api/register                → Открыт (без токена)
/api/login                   → Открыт (без токена)
/api/rates                   → Открыт (без токена)

/api/accounts                → ЗАЩИЩЁН (нужен токен)
/api/accounts/{id}           → ЗАЩИЩЁН
/api/accounts/{id}/transactions → ЗАЩИЩЁН
```

В `main.go` это выглядит так:

```go
// Открытые
http.HandleFunc("/api/register", authHandler.HandleRegister)
http.HandleFunc("/api/login",    authHandler.HandleLogin)
http.HandleFunc("/api/rates",    rateHandler.Handle)

// Защищённые — обёрнуты в AuthMiddleware
http.HandleFunc("/api/accounts",  handler.AuthMiddleware(accountHandler.HandleList))
http.HandleFunc("/api/accounts/", handler.AuthMiddleware(func(w, r) { ... }))
```

---

## 4. Backend — привязка данных к пользователю

### Что изменилось в SQL-запросах

Ключевое изменение: везде добавлен `WHERE user_id = $X`.

**Было** — список всех счетов (для всех):

```sql
SELECT ... FROM accounts ORDER BY id
```

**Стало** — только счета текущего пользователя:

```sql
SELECT ... FROM accounts WHERE user_id = $1 ORDER BY id
```

### Полная таблица изменений

| Метод             | Было                          | Стало                                |
|-------------------|-------------------------------|--------------------------------------|
| `GetAll()`        | Все счета                     | `GetAll(userID)` — только свои       |
| `GetByID(id)`     | Любой счёт по id              | `GetByID(id, userID)` — только свой  |
| `Create(account)` | Без привязки                  | `account.UserID = userID` — привязка |
| `Delete(id)`      | Любой счёт                    | `Delete(id, userID)` — только свой   |
| `Exists(id)`      | Любой счёт                    | `Exists(id, userID)` — только свой   |

### Пример: защита от доступа к чужим данным

```
Иван (user_id=1) пытается удалить счёт Анны (id=3, user_id=2):

  DELETE /api/accounts/3
  Authorization: Bearer <токен Ивана, user_id=1>

  1. Middleware: user_id = 1
  2. Handler:   Delete(id=3, userID=1)
  3. SQL:       DELETE FROM accounts WHERE id = 3 AND user_id = 1
  4. Результат: 0 строк удалено (у Ивана нет счёта с id=3)
  5. Ответ:     404 "Счёт не найден"

  ✅ Счёт Анны в безопасности
```

### То же самое с транзакциями

Перед добавлением транзакции проверяется, что счёт принадлежит пользователю:

```go
// handler/transaction.go
exists, _ := h.accountUC.Exists(accountID, userID)
//                                         ^^^^^^ из JWT-токена

if !exists {
    // 404 — счёт не найден (или не ваш)
}
```

---

## 5. Frontend — авторизация

### Хранение токена

После успешного входа токен сохраняется в `localStorage` браузера:

```
localStorage
  └── "token" → "eyJhbGciOiJIUzI1NiJ9..."
```

`localStorage` — это хранилище в браузере. Данные сохраняются даже после закрытия вкладки.

**Файл:** `src/stores/auth.ts`

```typescript
export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('token'))  // при старте читаем из localStorage

    function setToken(newToken: string) {
        token.value = newToken
        localStorage.setItem('token', newToken)        // сохраняем
    }

    function logout() {
        token.value = null
        localStorage.removeItem('token')               // удаляем
    }

    const isAuthenticated = computed(() => !!token.value)  // есть токен = залогинен
})
```

### `apiFetch` — обёртка над fetch

Раньше все запросы были простым `fetch()`. Теперь нужно к каждому добавлять заголовок с токеном. Чтобы не копировать это в каждом файле, создана обёртка:

**Файл:** `src/api.ts`

```
  ┌──────────────────────────────────────────────────┐
  │  apiFetch('/api/accounts')                        │
  │                                                    │
  │  Автоматически:                                    │
  │  1. Берёт токен из auth store                      │
  │  2. Добавляет заголовок:                           │
  │     Authorization: Bearer eyJhbGci...              │
  │  3. Если ответ 401 (токен истёк):                  │
  │     → Удаляет токен                                │
  │     → Перенаправляет на /login                     │
  └──────────────────────────────────────────────────┘
```

**Было** (без авторизации):

```typescript
fetch('/api/accounts')
fetch('/api/accounts', { method: 'POST', body: '...' })
```

**Стало** (с авторизацией):

```typescript
apiFetch('/api/accounts')
apiFetch('/api/accounts', { method: 'POST', body: '...' })
```

Просто заменили `fetch` на `apiFetch` — всё остальное делается автоматически.

### Router guard — защита страниц

Роутер (vue-router) теперь проверяет авторизацию перед каждым переходом:

```
┌─────────────────────────────────────────────┐
│  router.beforeEach()                         │
│                                              │
│  Куда идёт пользователь?                     │
│                                              │
│  /login или /register                        │
│    └── Уже залогинен? → Перекидываем на      │
│        /accounts (незачем логиниться снова)   │
│    └── Не залогинен? → Показываем форму      │
│                                              │
│  /accounts или /accounts/:id                 │
│    └── Залогинен? → Показываем страницу      │
│    └── Не залогинен? → Перекидываем на /login │
│                                              │
│  / (корень)                                  │
│    └── Всегда перекидываем на /accounts       │
└─────────────────────────────────────────────┘
```

**Файл:** `src/router/index.ts`

Маршруты с `meta: { requiresAuth: true }` — защищённые:

```typescript
{
  path: '/accounts',
  meta: { requiresAuth: true },   // ← эта метка включает проверку
  component: () => import('AccountsView.vue'),
}
```

---

## 6. Frontend — редизайн

### Что удалено (дефолтный шаблон Vue)

```
УДАЛЕНО:
  src/components/HelloWorld.vue       ← "You did it!"
  src/components/TheWelcome.vue       ← Приветственный блок
  src/components/WelcomeItem.vue      ← Элемент приветствия
  src/components/icons/Icon*.vue      ← 5 иконок Vue
  src/views/HomeView.vue              ← Главная страница
  src/views/AboutView.vue             ← Страница "О нас"
  src/assets/logo.svg                 ← Логотип Vue
  src/stores/counter.ts               ← Пример Pinia-стора
  src/layouts/LightLayout.vue         ← Старый лейаут
```

### Как выглядит приложение теперь

**Страница входа** (`/login`):

```
┌─────────────────────────────────────────────┐
│                                             │
│   (тёмный градиентный фон)                  │
│                                             │
│          ┌─────────────────────┐            │
│          │     FinTrack         │            │
│          │  Управление финансами│            │
│          │                     │            │
│          │  Вход               │            │
│          │                     │            │
│          │  Email              │            │
│          │  ┌─────────────────┐│            │
│          │  │                 ││            │
│          │  └─────────────────┘│            │
│          │  Пароль             │            │
│          │  ┌─────────────────┐│            │
│          │  │                 ││            │
│          │  └─────────────────┘│            │
│          │                     │            │
│          │  ┌─────────────────┐│            │
│          │  │     Войти        ││            │
│          │  └─────────────────┘│            │
│          │                     │            │
│          │  Нет аккаунта?      │            │
│          │  Зарегистрироваться │            │
│          └─────────────────────┘            │
│                                             │
└─────────────────────────────────────────────┘
```

**Главная страница** (`/accounts`, после входа):

```
┌─────────────────────────────────────────────────┐
│  FinTrack     Счета                     [Выйти] │ ← тёмно-синий topbar
├─────────────────────────────────────────────────┤
│                                                  │
│  Мои счета                                       │
│                                                  │
│  ┌──────────────────────────┐  ┌──────────────┐ │
│  │ Новый счёт               │  │ Итого        │ │
│  │ [Валюта][Коммент][Создать]│  │              │ │
│  └──────────────────────────┘  │ USD  1500.00 │ │
│                                 │ EUR   800.00 │ │
│  ┌──────────────────────────┐  │              │ │
│  │ USD  Основной    1500.00 │  │──────────────│ │
│  │ EUR  Отпуск       800.00 │  │≈USD  2834.50│ │
│  │ RUB  Зарплата  50000.00  │  └──────────────┘ │
│  └──────────────────────────┘                    │
│                                                  │
│  ← основная колонка →   ← боковая (итого) →      │
└─────────────────────────────────────────────────┘
```

**Было:**
- Логотип Vue наверху
- "You did it!" текст
- Навигация: Home | About | Счета
- Одна колонка на всю ширину
- LightLayout с белой шапкой

**Стало:**
- Тёмно-синий topbar с "FinTrack" и кнопкой "Выйти"
- Двухколоночная сетка: слева — счета, справа — итого
- Без топбара на страницах входа/регистрации
- Карточки с мягкими тенями

### Цветовая схема

| Элемент             | Цвет        | Hex       |
|---------------------|-------------|-----------|
| Topbar              | Тёмно-синий | `#0f3460` |
| Кнопки / ссылки     | Тёмно-синий | `#0f3460` |
| Доход / баланс +    | Зелёный     | `#22863a` |
| Расход / баланс -   | Красный     | `#d73a49` |
| Фон страницы        | Светло-серый| `#f5f6fa` |
| Карточки            | Белый       | `#ffffff` |
| Фон авторизации     | Градиент    | `#1a1a2e → #0f3460` |

---

## 7. Курсы валют

| Параметр            | Было       | Стало       |
|---------------------|------------|-------------|
| Обновление на сервере | 5 минут   | **1 час**  |
| Опрос с фронтенда    | 5 секунд  | **60 секунд** |

**Файл:** `internal/usecase/rate.go`

```go
// Было:
ticker := time.NewTicker(5 * time.Minute)

// Стало:
ticker := time.NewTicker(1 * time.Hour)
```

**Файл:** `src/views/AccountsView.vue`

```typescript
// Было:
setInterval(loadRates, 5000)   // каждые 5 секунд

// Стало:
setInterval(loadRates, 60000)  // каждые 60 секунд
```

Зачем: бесплатный план exchangerate-api.com имеет лимит на количество запросов. Раз в час — более чем достаточно, курсы не меняются каждые 5 минут.

---

## 8. Примеры запросов (curl)

### 1. Регистрация

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@mail.com", "password": "secret123"}'
```

Ответ `201 Created`:

```json
{
  "id": 1,
  "email": "ivan@mail.com",
  "created_at": "2026-02-19T12:00:00Z"
}
```

### 2. Вход

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@mail.com", "password": "secret123"}'
```

Ответ `200 OK`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 3. Сохраняем токен в переменную (для удобства)

```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. Создать счёт (с токеном)

```bash
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"currency": "USD", "comment": "Основной"}'
```

Ответ `201 Created`:

```json
{
  "id": 1,
  "user_id": 1,
  "currency": "USD",
  "comment": "Основной",
  "created_at": "2026-02-19T12:05:00Z",
  "balance": 0
}
```

### 5. Список своих счетов

```bash
curl http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN"
```

Ответ:

```json
[
  {"id": 1, "user_id": 1, "currency": "USD", "comment": "Основной", "balance": 0}
]
```

### 6. Добавить операцию (пополнение)

```bash
curl -X POST http://localhost:8080/api/accounts/1/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"amount": 5000, "comment": "Зарплата"}'
```

Ответ `201 Created`:

```json
{
  "id": 1,
  "account_id": 1,
  "amount": 5000,
  "comment": "Зарплата",
  "created_at": "2026-02-19T12:10:00Z"
}
```

### 7. Запрос БЕЗ токена — ошибка

```bash
curl http://localhost:8080/api/accounts
```

Ответ `401 Unauthorized`:

```json
{"error": "Требуется авторизация"}
```

### 8. Неверный пароль — ошибка

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@mail.com", "password": "wrong"}'
```

Ответ `401 Unauthorized`:

```json
{"error": "Неверный email или пароль"}
```

### 9. Повторная регистрация с тем же email — ошибка

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"email": "ivan@mail.com", "password": "another"}'
```

Ответ `409 Conflict`:

```json
{"error": "Пользователь с таким email уже существует"}
```

---

## 9. Карта файлов

### Новые файлы

```
BACKEND:
  db/migrations/000005_create_users_table.up.sql       ← таблица users
  db/migrations/000005_create_users_table.down.sql
  db/migrations/000006_add_user_id_to_accounts.up.sql  ← колонка user_id
  db/migrations/000006_add_user_id_to_accounts.down.sql

  internal/entity/user.go                    ← модель User
  internal/repository/postgres/user.go       ← Create, GetByEmail
  internal/usecase/auth.go                   ← Register (bcrypt), Login (JWT)
  internal/handler/auth.go                   ← POST /register, POST /login
  internal/handler/middleware.go             ← AuthMiddleware

FRONTEND:
  src/stores/auth.ts                         ← Pinia-стор (token, login, logout)
  src/api.ts                                 ← apiFetch() с Authorization header
  src/views/LoginView.vue                    ← Страница входа
  src/views/RegisterView.vue                 ← Страница регистрации
```

### Изменённые файлы

```
BACKEND:
  go.mod                                     ← добавлены jwt/v5, x/crypto
  cmd/app/main.go                            ← +userRepo, +authUC, +authHandler,
                                                маршруты обёрнуты в AuthMiddleware
  internal/entity/account.go                 ← +UserID поле
  internal/repository/postgres/account.go    ← все запросы фильтруют по user_id
  internal/usecase/account.go                ← все методы принимают userID
  internal/handler/account.go                ← берёт user_id из context
  internal/handler/transaction.go            ← проверяет принадлежность счёта
  internal/usecase/rate.go                   ← 5 min → 1 hour

FRONTEND:
  src/App.vue                                ← topbar FinTrack + logout
                                                (вместо логотипа Vue + HelloWorld)
  src/router/index.ts                        ← auth-маршруты + beforeEach guard
  src/views/AccountsView.vue                 ← 2-колоночная сетка, apiFetch
  src/views/AccountDetailView.vue            ← новый дизайн карточек, apiFetch
  src/assets/base.css                        ← убрана Vue-тема, минимальный reset
  src/assets/main.css                        ← упрощён
```

### Удалённые файлы

```
  src/components/HelloWorld.vue              ← "You did it!"
  src/components/TheWelcome.vue              ← Приветственный блок
  src/components/WelcomeItem.vue             ← Элемент приветствия
  src/components/icons/IconCommunity.vue     ← Иконки Vue
  src/components/icons/IconDocumentation.vue
  src/components/icons/IconEcosystem.vue
  src/components/icons/IconSupport.vue
  src/components/icons/IconTooling.vue
  src/views/HomeView.vue                     ← Главная страница (не нужна)
  src/views/AboutView.vue                    ← "О нас" (не нужна)
  src/assets/logo.svg                        ← Логотип Vue
  src/stores/counter.ts                      ← Пример стора (не используется)
  src/layouts/LightLayout.vue                ← Старый лейаут
```

---

## Как запустить и проверить

```bash
# 1. БД (если не запущена)
make db-up

# 2. Backend (миграции применятся автоматически)
make run

# 3. Frontend (в другом терминале)
npm run dev

# 4. Открыть http://localhost:5173
#    → Покажет страницу входа
#    → Перейти на "Зарегистрироваться"
#    → Ввести email + пароль → автоматический вход
#    → Создать счёт → Добавить операции
#    → Кнопка "Выйти" в правом верхнем углу
```

---

## Краткая сводка (10 пунктов)

1. **Таблица `users`** — email + bcrypt-хеш пароля
2. **Колонка `user_id` в `accounts`** — привязка счёта к владельцу
3. **Регистрация** (`POST /api/register`) — хешируем пароль, сохраняем
4. **Вход** (`POST /api/login`) — проверяем хеш, выдаём JWT на 72 часа
5. **JWT Middleware** — проверяет токен, кладёт `user_id` в context
6. **Все запросы к счетам** — фильтруются по `user_id` из токена
7. **Фронтенд** — токен в localStorage, `apiFetch()` добавляет заголовок
8. **Router guard** — без токена → редирект на `/login`
9. **Новый UI** — topbar "FinTrack", карточки, двухколоночная сетка
10. **Курсы** — обновляются раз в час (было 5 минут)
