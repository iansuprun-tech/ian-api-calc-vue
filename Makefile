# Загружаем переменные из .env файла (если он есть)
# include — директива Make, она читает другой файл и подставляет переменные.
# "-" перед include означает "не падай если файла нет".
-include .env

# ?= означает "использовать это значение, только если переменная ещё не задана".
# Если DATABASE_URL уже загружен из .env — он не перезапишется.
DB_URL ?= postgres://postgres:postgres@localhost:5432/balances?sslmode=disable
MIGRATIONS_DIR = db/migrations

# Запустить PostgreSQL через docker-compose
.PHONY: db-up
db-up:
	docker-compose up -d

# Остановить PostgreSQL
.PHONY: db-down
db-down:
	docker-compose down

# Применить все миграции
.PHONY: migrate-up
migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

# Откатить последнюю миграцию
.PHONY: migrate-down
migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

# Откатить все миграции
.PHONY: migrate-down-all
migrate-down-all:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down -all

# Текущая версия миграций
.PHONY: migrate-version
migrate-version:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) version

# Создать новую миграцию. Использование: make migrate-create name=add_users_table
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then echo "Ошибка: укажи имя. Пример: make migrate-create name=add_users_table"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Исправить dirty state. Использование: make migrate-force v=1
.PHONY: migrate-force
migrate-force:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) force $(v)

# Запустить Go-сервер
.PHONY: run
run:
	go run ./cmd/app/
