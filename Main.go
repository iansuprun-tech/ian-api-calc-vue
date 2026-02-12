package main

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

	_ "github.com/mattn/go-sqlite3" // SQLite драйвер
)

const exchangeRateAPIKey = "6d261a66cfd0da7dfd5597e6"
const exchangeRateAPIURL = "https://v6.exchangerate-api.com/v6/"

type Balance struct {
	ID       int     `json:"id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type Rates struct {
	ID        int     `json:"id"`
	Currency  string  `json:"currency"`
	RateToUSD float64 `json:"rate_to_usd"`
	UpdateAt  string  `json:"update_at"`
}

type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

// Глобальная переменная для подключения к БД
var db *sql.DB

// TODO: зачем нужны пустые круглые скобки? (так и не разобрался)
func main() {
	// Инициализация базы данных
	//TODO: что такое "err"?
	var err error
	db, err = sql.Open("sqlite3", "./balances.db")
	if err != nil {
		log.Fatal("Ошибка подключения к БД", err)
	}
	defer db.Close()

	// Создание таблицы задач
	createTable()

	// Настройка маршрутов
	http.HandleFunc("/api/balances", handleBalancesList)
	http.HandleFunc("/api/balances/", handleBalanceById)

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

// CreateTable создает таблицу task если ее нет
func createTable() {
	//TODO: зачем нужно ":=" "!=" и тд
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

func fetchAndSaveRates() {
	url := exchangeRateAPIURL + exchangeRateAPIKey + "/latest/USD"
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
INSERT INTO rates (currency, rate_to_usd, update_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (currency)
		DO UPDATE SET rate_to_usd = ?, update_at = CURRENT_TIMESTAMP`,
			currency, rateToUSD, rateToUSD)

		if err != nil {
			log.Println("Ошибка сохранения курса для", currency, ":", err)
		}
	}

	log.Println("Курсы валют обновлены успешно!")
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
	rows, err := db.Query("SELECT id, currency, amount, rate FROM balances")
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
		if err := rows.Scan(&balance.ID, &balance.Currency, &balance.Amount, &balance.Rate); err != nil {
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

	result, err := db.Exec("INSERT INTO balances (currency, amount, rate) VALUES (?, ?, ?)",
		balance.Currency, balance.Amount, balance.Rate)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания задачи"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	balance.ID = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(balance)
}

// getBalance возвращает задачу по ID
func getBalance(w http.ResponseWriter, r *http.Request, id int) {
	var balance Balance
	err := db.QueryRow("SELECT id, currency, amount, rate FROM balances WHERE id = ?", id).
		Scan(&balance.ID, &balance.Currency, &balance.Amount, &balance.Rate)

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
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	}

	_, err = db.Exec("UPDATE balances SET currency = ?, amount = ?, rate = ? WHERE id = ?",
		balance.Currency, balance.Amount, balance.Rate, id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка обновления задачи"}`, http.StatusInternalServerError)
		return
	}

	balance.ID = id
	json.NewEncoder(w).Encode(balance)
}

// deleteBalance удаляет задачу по ID
func deleteBalance(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec("DELETE FROM balances WHERE id = ?", id)
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
