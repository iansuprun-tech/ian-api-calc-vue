// Точка входа в приложение.
// Здесь собираем все зависимости вместе: БД -> Репозитории -> Юзкейсы -> Хендлеры.
// Этот файл — единственное место, которое знает обо всех слоях.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	dbpkg "vue-calc/db"
	"vue-calc/internal/entity"
	"vue-calc/internal/handler"
	"vue-calc/internal/repository/postgres"
	"vue-calc/internal/usecase"
)

// exchangeRateAPIURL — базовый URL внешнего API для получения курсов валют.
const exchangeRateAPIURL = "https://v6.exchangerate-api.com/v6/"

// rateFetcher — реализация usecase.RateFetcher через HTTP-запрос к exchangerate-api.com.
type rateFetcher struct {
	apiKey string
}

// FetchRates делает HTTP-запрос к API и возвращает распарсенный ответ.
func (f *rateFetcher) FetchRates() (*entity.ExchangeRateResponse, error) {
	if f.apiKey == "" {
		log.Println("EXCHANGE_RATE_API_KEY не задан, пропускаем обновление курсов")
		return &entity.ExchangeRateResponse{Result: "skip"}, nil
	}

	url := exchangeRateAPIURL + f.apiKey + "/latest/USD"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к API курсов: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа API: %w", err)
	}

	var rateResponse entity.ExchangeRateResponse
	if err := json.Unmarshal(body, &rateResponse); err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON от API: %w", err)
	}

	return &rateResponse, nil
}

func main() {
	// Загружаем переменные из .env файла.
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные окружения")
	}

	// Подключение к БД
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/balances?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("БД недоступна: ", err)
	}
	defer db.Close()

	// Применение миграций
	runMigrations(dsn)

	// --- Сборка зависимостей (Dependency Injection) ---
	// 1. Создаём репозитории (слой данных)
	accountRepo := postgres.NewAccountRepo(db)
	transactionRepo := postgres.NewTransactionRepo(db)
	categoryRepo := postgres.NewCategoryRepo(db)
	rateRepo := postgres.NewRateRepo(db)
	userRepo := postgres.NewUserRepo(db)

	// 2. Создаём юзкейсы (бизнес-логика), передавая им репозитории
	accountUC := usecase.NewAccountUseCase(accountRepo)
	transactionUC := usecase.NewTransactionUseCase(transactionRepo)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	fetcher := &rateFetcher{apiKey: os.Getenv("EXCHANGE_RATE_API_KEY")}
	rateUC := usecase.NewRateUseCase(rateRepo, fetcher)
	authUC := usecase.NewAuthUseCase(userRepo)

	// 3. Создаём хендлеры (HTTP-слой), передавая им юзкейсы
	accountHandler := handler.NewAccountHandler(accountUC)
	transactionHandler := handler.NewTransactionHandler(transactionUC, accountUC)
	categoryHandler := handler.NewCategoryHandler(categoryUC)
	rateHandler := handler.NewRateHandler(rateUC)
	authHandler := handler.NewAuthHandler(authUC)

	// Запускаем фоновое обновление курсов валют
	rateUC.StartUpdater()

	// Публичные маршруты (без авторизации)
	http.HandleFunc("/api/register", authHandler.HandleRegister)
	http.HandleFunc("/api/login", authHandler.HandleLogin)
	http.HandleFunc("/api/rates", rateHandler.Handle)

	// Защищённые маршруты (требуют JWT)
	http.HandleFunc("/api/categories", handler.AuthMiddleware(categoryHandler.Handle))
	http.HandleFunc("/api/categories/", handler.AuthMiddleware(categoryHandler.Handle))
	http.HandleFunc("/api/accounts", handler.AuthMiddleware(accountHandler.HandleList))
	http.HandleFunc("/api/accounts/", handler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/accounts/") && strings.Contains(path, "/transactions") {
			transactionHandler.Handle(w, r)
		} else {
			accountHandler.HandleByID(w, r)
		}
	}))

	// Запуск сервера
	fmt.Println("Сервер запущен на http://localhost:8080")
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /api/register                   - регистрация")
	fmt.Println("  POST   /api/login                      - вход")
	fmt.Println("  GET    /api/accounts                    - список всех счетов")
	fmt.Println("  POST   /api/accounts                    - создать счёт")
	fmt.Println("  GET    /api/accounts/{id}               - получить счёт")
	fmt.Println("  DELETE /api/accounts/{id}               - удалить счёт")
	fmt.Println("  GET    /api/accounts/{id}/transactions  - история операций")
	fmt.Println("  POST   /api/accounts/{id}/transactions  - добавить операцию")
	fmt.Println("  GET    /api/categories                   - список категорий")
	fmt.Println("  POST   /api/categories                   - создать категорию")
	fmt.Println("  DELETE /api/categories/{id}               - удалить категорию")
	fmt.Println("  GET    /api/rates                       - список курсов валют")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// runMigrations применяет все pending миграции из встроенных SQL-файлов.
func runMigrations(dsn string) {
	sourceDriver, err := iofs.New(dbpkg.MigrationsFS, "migrations")
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
