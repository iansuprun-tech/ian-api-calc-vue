package entity

// CategoryStat — агрегированная статистика по одной категории.
type CategoryStat struct {
	CategoryID   *int    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
}

// DailyStat — доходы и расходы за один день (для bar chart).
type DailyStat struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// CurrencyStats — статистика по одной валюте.
type CurrencyStats struct {
	Currency          string         `json:"currency"`
	TotalIncome       float64        `json:"total_income"`
	TotalExpense      float64        `json:"total_expense"`
	IncomeByCategory  []CategoryStat `json:"income_by_category"`
	ExpenseByCategory []CategoryStat `json:"expense_by_category"`
	DailyStats        []DailyStat    `json:"daily_stats"`
}

// StatisticsResponse — полный ответ со статистикой за период, разбитый по валютам.
type StatisticsResponse struct {
	Currencies []CurrencyStats `json:"currencies"`
}
