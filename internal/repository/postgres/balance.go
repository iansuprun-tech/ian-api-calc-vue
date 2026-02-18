// Пакет postgres содержит реализации репозиториев для PostgreSQL.
// Здесь живут все SQL-запросы — остальной код про них не знает.
package postgres

import (
	"database/sql"

	"vue-calc/internal/entity"
)

// BalanceRepo — реализация BalanceRepository для PostgreSQL.
// Хранит ссылку на подключение к БД.
type BalanceRepo struct {
	db *sql.DB
}

// NewBalanceRepo — конструктор.
func NewBalanceRepo(db *sql.DB) *BalanceRepo {
	return &BalanceRepo{db: db}
}

// GetAll возвращает все балансы из таблицы balances.
func (r *BalanceRepo) GetAll() ([]entity.Balance, error) {
	rows, err := r.db.Query("SELECT id, currency, amount FROM balances")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := []entity.Balance{}
	for rows.Next() {
		var b entity.Balance
		if err := rows.Scan(&b.ID, &b.Currency, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}

// GetByID возвращает баланс по ID. Возвращает sql.ErrNoRows, если не найден.
func (r *BalanceRepo) GetByID(id int) (entity.Balance, error) {
	var b entity.Balance
	err := r.db.QueryRow("SELECT id, currency, amount FROM balances WHERE id = $1", id).
		Scan(&b.ID, &b.Currency, &b.Amount)
	return b, err
}

// Create вставляет новый баланс и возвращает его с присвоенным ID.
func (r *BalanceRepo) Create(balance entity.Balance) (entity.Balance, error) {
	err := r.db.QueryRow(
		"INSERT INTO balances (currency, amount) VALUES ($1, $2) RETURNING id",
		balance.Currency, balance.Amount,
	).Scan(&balance.ID)
	return balance, err
}

// Update обновляет валюту и сумму баланса по ID.
func (r *BalanceRepo) Update(id int, balance entity.Balance) error {
	_, err := r.db.Exec(
		"UPDATE balances SET currency = $1, amount = $2 WHERE id = $3",
		balance.Currency, balance.Amount, id,
	)
	return err
}

// Delete удаляет баланс по ID. Возвращает количество затронутых строк.
func (r *BalanceRepo) Delete(id int) (int64, error) {
	result, err := r.db.Exec("DELETE FROM balances WHERE id = $1", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Exists проверяет, существует ли баланс с данным ID.
func (r *BalanceRepo) Exists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
