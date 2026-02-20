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
// from/to — даты в формате "2025-01-01", accountID — опциональный фильтр по счёту.
func (r *StatisticsRepo) GetStatistics(userID int, from, to string, accountID *int) (entity.StatisticsResponse, error) {
	var result entity.StatisticsResponse

	// Суммы доходов и расходов
	totals, err := r.getTotals(userID, from, to, accountID)
	if err != nil {
		return result, err
	}
	result.TotalIncome = totals.income
	result.TotalExpense = totals.expense

	// Разбивка по категориям — доходы
	result.IncomeByCategory, err = r.getCategoryStats(userID, from, to, accountID, true)
	if err != nil {
		return result, err
	}

	// Разбивка по категориям — расходы
	result.ExpenseByCategory, err = r.getCategoryStats(userID, from, to, accountID, false)
	if err != nil {
		return result, err
	}

	// Статистика по дням
	result.DailyStats, err = r.getDailyStats(userID, from, to, accountID)
	if err != nil {
		return result, err
	}

	return result, nil
}

type totalsResult struct {
	income  float64
	expense float64
}

func (r *StatisticsRepo) getTotals(userID int, from, to string, accountID *int) (totalsResult, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN t.amount > 0 THEN t.amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.amount < 0 THEN ABS(t.amount) ELSE 0 END), 0) AS expense
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

	var res totalsResult
	err := r.db.QueryRow(query, args...).Scan(&res.income, &res.expense)
	return res, err
}

func (r *StatisticsRepo) getCategoryStats(userID int, from, to string, accountID *int, isIncome bool) ([]entity.CategoryStat, error) {
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
		  AND t.created_at >= $2
		  AND t.created_at < ($3::date + interval '1 day')
		  AND ` + amountCondition

	args := []interface{}{userID, from, to}

	if accountID != nil {
		query += " AND t.account_id = $4"
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

func (r *StatisticsRepo) getDailyStats(userID int, from, to string, accountID *int) ([]entity.DailyStat, error) {
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
		  AND t.created_at >= $2
		  AND t.created_at < ($3::date + interval '1 day')`

	args := []interface{}{userID, from, to}

	if accountID != nil {
		query += " AND t.account_id = $4"
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
