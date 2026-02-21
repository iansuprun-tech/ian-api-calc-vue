package postgres

import (
	"database/sql"
	"vue-calc/internal/entity"
)

// StatisticsRepo — репозиторий для агрегации статистики по транзакциям.
type StatisticsRepo struct {
	db *sql.DB
}

// NewStatisticsRepo — конструктор репозитория статистики.
func NewStatisticsRepo(db *sql.DB) *StatisticsRepo {
	return &StatisticsRepo{db: db}
}

// GetStatistics — получить агрегированную статистику за период, сгруппированную по валютам.
func (r *StatisticsRepo) GetStatistics(userID int, from, to string, accountID *int) (entity.StatisticsResponse, error) {
	var result entity.StatisticsResponse

	// Получаем список валют с транзакциями за период
	currencies, err := r.getCurrencies(userID, from, to, accountID)
	if err != nil {
		return result, err
	}

	for _, currency := range currencies {
		cs := entity.CurrencyStats{Currency: currency}

		totals, err := r.getTotals(userID, from, to, accountID, currency)
		if err != nil {
			return result, err
		}
		cs.TotalIncome = totals.income
		cs.TotalExpense = totals.expense

		cs.IncomeByCategory, err = r.getCategoryStats(userID, from, to, accountID, currency, true)
		if err != nil {
			return result, err
		}

		cs.ExpenseByCategory, err = r.getCategoryStats(userID, from, to, accountID, currency, false)
		if err != nil {
			return result, err
		}

		cs.DailyStats, err = r.getDailyStats(userID, from, to, accountID, currency)
		if err != nil {
			return result, err
		}

		result.Currencies = append(result.Currencies, cs)
	}

	if result.Currencies == nil {
		result.Currencies = []entity.CurrencyStats{}
	}

	return result, nil
}

// getCurrencies — получить список валют, по которым есть транзакции за период.
func (r *StatisticsRepo) getCurrencies(userID int, from, to string, accountID *int) ([]string, error) {
	query := `
		SELECT DISTINCT a.currency
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND t.created_at >= $2
		  AND t.created_at < ($3::date + interval '1 day')`

	args := []interface{}{userID, from, to}

	if accountID != nil {
		query += " AND t.account_id = $4"
		args = append(args, *accountID)
	}

	query += " ORDER BY a.currency"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, nil
}

type totalsResult struct {
	income  float64
	expense float64
}

func (r *StatisticsRepo) getTotals(userID int, from, to string, accountID *int, currency string) (totalsResult, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) ELSE 0 END), 0) AS expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND a.currency = $2
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')`

	args := []interface{}{userID, currency, from, to}

	if accountID != nil {
		query += " AND t.account_id = $5"
		args = append(args, *accountID)
	}

	var res totalsResult
	err := r.db.QueryRow(query, args...).Scan(&res.income, &res.expense)
	return res, err
}

func (r *StatisticsRepo) getCategoryStats(userID int, from, to string, accountID *int, currency string, isIncome bool) ([]entity.CategoryStat, error) {
	amountCondition := "t.amount > 0"
	sumExpr := "COALESCE(SUM(t.amount), 0)"
	if !isIncome {
		amountCondition = "t.amount < 0"
		sumExpr = "COALESCE(SUM(ABS(t.amount)), 0)"
	}

	query := `
		SELECT t.category_id, COALESCE(c.name, 'Без категории'), ` + sumExpr + `, COUNT(*)
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND a.currency = $2
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')
		  AND ` + amountCondition

	args := []interface{}{userID, currency, from, to}

	if accountID != nil {
		query += " AND t.account_id = $5"
		args = append(args, *accountID)
	}

	query += " GROUP BY t.category_id, c.name ORDER BY SUM(ABS(t.amount)) DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.CategoryStat
	for rows.Next() {
		var s entity.CategoryStat
		if err := rows.Scan(&s.CategoryID, &s.CategoryName, &s.Total, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []entity.CategoryStat{}
	}
	return stats, nil
}

func (r *StatisticsRepo) getDailyStats(userID int, from, to string, accountID *int, currency string) ([]entity.DailyStat, error) {
	query := `
		SELECT
			DATE(t.created_at)::text AS day,
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) ELSE 0 END), 0) AS expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND a.currency = $2
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')`

	args := []interface{}{userID, currency, from, to}

	if accountID != nil {
		query += " AND t.account_id = $5"
		args = append(args, *accountID)
	}

	query += " GROUP BY DATE(t.created_at) ORDER BY DATE(t.created_at)"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []entity.DailyStat
	for rows.Next() {
		var s entity.DailyStat
		if err := rows.Scan(&s.Date, &s.Income, &s.Expense); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []entity.DailyStat{}
	}
	return stats, nil
}
