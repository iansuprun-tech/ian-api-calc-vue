package postgres

import (
	"database/sql"

	"vue-calc/internal/entity"
)

// RateRepo — реализация RateRepository для PostgreSQL.
type RateRepo struct {
	db *sql.DB
}

// NewRateRepo — конструктор.
func NewRateRepo(db *sql.DB) *RateRepo {
	return &RateRepo{db: db}
}

// GetAll возвращает все курсы валют из таблицы rates.
func (r *RateRepo) GetAll() ([]entity.Rate, error) {
	rows, err := r.db.Query("SELECT id, currency, rate_to_usd, updated_at FROM rates")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rates := []entity.Rate{}
	for rows.Next() {
		var rate entity.Rate
		if err := rows.Scan(&rate.ID, &rate.Currency, &rate.RateToUSD, &rate.UpdatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

// Upsert вставляет или обновляет курс валюты (INSERT ... ON CONFLICT DO UPDATE).
// Если валюта уже есть — обновляет курс и время.
func (r *RateRepo) Upsert(currency string, rateToUSD float64) error {
	_, err := r.db.Exec(`
		INSERT INTO rates (currency, rate_to_usd, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (currency)
		DO UPDATE SET rate_to_usd = $2, updated_at = CURRENT_TIMESTAMP`,
		currency, rateToUSD)
	return err
}
