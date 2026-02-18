package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv" // загружает переменные из .env файла
	_ "github.com/lib/pq"      // PostgreSQL драйвер
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

const exchangeRateAPIURL = "https://v6.exchangerate-api.com/v6/"

type Balance struct {
	ID       int     `json:"id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type Rate struct {
	ID        int     `json:"id"`
	Currency  string  `json:"currency"`
	RateToUSD float64 `json:"rate_to_usd"`
	UpdatedAt string  `json:"updated_at"`
}

type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

// Глобальная переменная для подключения к БД
var db *sql.DB

// TODO: зачем нужны пустые круглые скобки? (так и не разобрался)
func main() {
	// Загружаем переменные из .env файла.
	// godotenv.Load() читает файл .env и добавляет все переменные в окружение,
	// после чего их можно получать через os.Getenv("ИМЯ_ПЕРЕМЕННОЙ").
	// Если .env нет (например, в продакшене) — не падаем, а используем системные переменные.
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные окружения")
	}

	// Инициализация базы данных
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/balances?sslmode=disable"
	}
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("БД недоступна: ", err)
	}
	defer db.Close()

	// Применение миграций
	runMigrations(dsn)

	startRateUpdater()

	// Настройка маршрутов
	http.HandleFunc("/api/balances", handleBalancesList)
	http.HandleFunc("/api/balances/", handleBalanceById)

	http.HandleFunc("/api/rates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getRates(w, r)
		} else {
			http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		}
	})

	// Запуск сервера
	fmt.Println("Сервер запущен на http//localhost:8080")
	fmt.Println("Endpoints")
	fmt.Println(" GET /api/balances - список всех задач")
	fmt.Println(" POST /api/balances - создать задачу")
	fmt.Println(" GET /api/balances/{id} - получить задачу")
	fmt.Println(" PUT /api/balances/{id} - обновить задачу")
	fmt.Println(" DELETE ?api/balances/{id}} - удалить задачу")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// runMigrations применяет все pending миграции из встроенных SQL-файлов
func runMigrations(dsn string) {
	sourceDriver, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		log.Fatal("Ошибка инициализации миграций: ", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		log.Fatal("Ошибка создания мигратора: ", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Ошибка применения миграций: ", err)
	}

	log.Println("Миграции применены успешно!")
}

func fetchAndSaveRates() {
	apiKey := os.Getenv("EXCHANGE_RATE_API_KEY")
	if apiKey == "" {
		log.Println("EXCHANGE_RATE_API_KEY не задан, пропускаем обновление курсов")
		return
	}
	url := exchangeRateAPIURL + apiKey + "/latest/USD"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Ошибка запроса к API курсов:", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения ответа API:", err)
		return
	}

	var rateResponse ExchangeRateResponse
	err = json.Unmarshal(body, &rateResponse)
	if err != nil {
		log.Println("Ошибка разбора JSON от API:", err)
		return
	}

	if rateResponse.Result != "success" {
		log.Println("API вернул ошибку, result:", rateResponse.Result)
		return
	}

	for currency, rate := range rateResponse.ConversionRates {
		var rateToUSD float64
		if rate > 0 {
			rateToUSD = 1.0 / rate
		}

		_, err := db.Exec(`
INSERT INTO rates (currency, rate_to_usd, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (currency)
		DO UPDATE SET rate_to_usd = $2, updated_at = CURRENT_TIMESTAMP`,
			currency, rateToUSD)

		if err != nil {
			log.Println("Ошибка сохранения курса для", currency, ":", err)
		}
	}

	log.Println("Курсы валют обновлены успешно!")
}

func startRateUpdater() {
	fetchAndSaveRates()
	ticker := time.NewTicker(5 * time.Minute)

	go func() {
		for range ticker.C {
			fetchAndSaveRates()
		}
	}()
}

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
			http.Error(w, `{"error": "Ошилбка чтения курсов"}`, http.StatusInternalServerError)
			return
		}
		rates = append(rates, rate)
	}

	json.NewEncoder(w).Encode(rates)
}

// handleBalances обрабатывает /balances (GET - список, POST - создание)
func handleBalancesList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getBalances(w, r)
	case http.MethodPost:
		createBalance(w, r)
	default:
		http.Error(w, "{\"error\": \"Метод не поддерживается\"}", http.StatusMethodNotAllowed)
	}
}

// handleBalanceByID обрабатывает /balances/{id} (GET, PUT, DELETE)
func handleBalanceById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Извлекаем ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/balances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Неверный ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getBalance(w, r, id)
	case http.MethodPut:

		updateBalance(w, r, id)
	case http.MethodDelete:

		deleteBalance(w, r, id)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

func getBalances(w http.ResponseWriter, _ *http.Request) {
	//TODO: что такое rows
	//TODO: rows - это строка
	rows, err := db.Query("SELECT id, currency, amount FROM balances")
	//TODO: зачем нужен капсовый текст
	//TODO: капсовый текст - название действия
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения задач"}`, http.StatusInternalServerError)
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

// CreateBalance создает новую задачу
func createBalance(w http.ResponseWriter, r *http.Request) {
	var balance Balance
	if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	err := db.QueryRow("INSERT INTO balances (currency, amount) VALUES ($1, $2) RETURNING id",
		balance.Currency, balance.Amount).Scan(&balance.ID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания задачи"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(balance)
}

// getBalance возвращает задачу по ID
func getBalance(w http.ResponseWriter, r *http.Request, id int) {
	var balance Balance
	err := db.QueryRow("SELECT id, currency, amount FROM balances WHERE id = $1", id).
		Scan(&balance.ID, &balance.Currency, &balance.Amount)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения задачи"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(balance)
}

// updateBalance обновляет задачу по ID
func updateBalance(w http.ResponseWriter, r *http.Request, id int) {
	var balance Balance
	if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
		http.Error(w, `{"error": "неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	_, err = db.Exec("UPDATE balances SET currency = $1, amount = $2 WHERE id = $3",
		balance.Currency, balance.Amount, id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка обновления задачи"}`, http.StatusInternalServerError)
		return
	}

	balance.ID = id
	json.NewEncoder(w).Encode(balance)
}

// deleteBalance удаляет задачу по ID
func deleteBalance(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec("DELETE FROM balances WHERE id = $1", id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка удаления задачи"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
