package entity

// Transaction — доменная модель операции (транзакции) по счёту.
// Положительное значение amount — пополнение, отрицательное — списание.
type Transaction struct {
	ID         int     `json:"id"`
	AccountID  int     `json:"account_id"`
	Amount     float64 `json:"amount"`
	Comment    string  `json:"comment"`
	CategoryID *int    `json:"category_id"`
	Category   string  `json:"category"`
	CreatedAt  string  `json:"created_at"`
}
