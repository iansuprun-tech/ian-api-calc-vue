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

// GetStatistics — получить агрегированную статистику за период.
// Все суммы пересчитываются в targetCurrency через таблицу rates.
func (r *StatisticsRepo) GetStatistics(userID int, from, to string, accountID *int, targetCurrency string) (entity.StatisticsResponse, error) {
	result := entity.StatisticsResponse{Currency: targetCurrency}

	totals, err := r.getTotals(userID, from, to, accountID, targetCurrency)
	if err != nil {
		return result, err
	}
	result.TotalIncome = totals.income
	result.TotalExpense = totals.expense

	result.IncomeByCategory, err = r.getCategoryStats(userID, from, to, accountID, targetCurrency, true)
	if err != nil {
		return result, err
	}

	result.ExpenseByCategory, err = r.getCategoryStats(userID, from, to, accountID, targetCurrency, false)
	if err != nil {
		return result, err
	}

	result.DailyStats, err = r.getDailyStats(userID, from, to, accountID, targetCurrency)
	if err != nil {
		return result, err
	}

	return result, nil
}

type totalsResult struct {
	income  float64
	expense float64
}

func (r *StatisticsRepo) getTotals(userID int, from, to string, accountID *int, targetCurrency string) (totalsResult, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount * (r_src.rate_to_usd / r_tgt.rate_to_usd) ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) * (r_src.rate_to_usd / r_tgt.rate_to_usd) ELSE 0 END), 0) AS expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		JOIN rates r_src ON r_src.currency = a.currency
		JOIN rates r_tgt ON r_tgt.currency = $2
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')`

	args := []interface{}{userID, targetCurrency, from, to}

	if accountID != nil {
		query += " AND t.account_id = $5"
		args = append(args, *accountID)
	}

	var res totalsResult
	err := r.db.QueryRow(query, args...).Scan(&res.income, &res.expense)
	return res, err
}

func (r *StatisticsRepo) getCategoryStats(userID int, from, to string, accountID *int, targetCurrency string, isIncome bool) ([]entity.CategoryStat, error) {
	amountCondition := "t.amount > 0"
	sumExpr := "COALESCE(SUM(t.amount * (r_src.rate_to_usd / r_tgt.rate_to_usd)), 0)"
	if !isIncome {
		amountCondition = "t.amount < 0"
		sumExpr = "COALESCE(SUM(ABS(t.amount) * (r_src.rate_to_usd / r_tgt.rate_to_usd)), 0)"
	}

	query := `
		SELECT t.category_id, COALESCE(c.name, 'Без категории'), ` + sumExpr + `, COUNT(*)
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		JOIN rates r_src ON r_src.currency = a.currency
		JOIN rates r_tgt ON r_tgt.currency = $2
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')
		  AND ` + amountCondition

	args := []interface{}{userID, targetCurrency, from, to}

	if accountID != nil {
		query += " AND t.account_id = $5"
		args = append(args, *accountID)
	}

	query += " GROUP BY t.category_id, c.name ORDER BY " + sumExpr + " DESC"

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

func (r *StatisticsRepo) getDailyStats(userID int, from, to string, accountID *int, targetCurrency string) ([]entity.DailyStat, error) {
	query := `
		SELECT
			DATE(t.created_at)::text AS day,
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount * (r_src.rate_to_usd / r_tgt.rate_to_usd) ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) * (r_src.rate_to_usd / r_tgt.rate_to_usd) ELSE 0 END), 0) AS expense
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		JOIN rates r_src ON r_src.currency = a.currency
		JOIN rates r_tgt ON r_tgt.currency = $2
		WHERE a.user_id = $1
		  AND t.deleted_at IS NULL
		  AND a.deleted_at IS NULL
		  AND t.created_at >= $3
		  AND t.created_at < ($4::date + interval '1 day')`

	args := []interface{}{userID, targetCurrency, from, to}

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
