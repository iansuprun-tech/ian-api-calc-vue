package entity

// Transaction — доменная модель операции (транзакции) по счёту.
// Положительное значение amount — пополнение, отрицательное — списание.
type Transaction struct {
	ID        int     `json:"id"`
	AccountID int     `json:"account_id"`
	Amount    float64 `json:"amount"`
	Comment   string  `json:"comment"`
	CreatedAt string  `json:"created_at"`
}
