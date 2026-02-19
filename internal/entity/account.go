package entity

// Account — доменная модель счёта.
// Счёт хранит валюту и комментарий. Баланс вычисляется как сумма всех транзакций по счёту.
type Account struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	Currency  string  `json:"currency"`
	Comment   string  `json:"comment"`
	CreatedAt string  `json:"created_at"`
	Balance   float64 `json:"balance"` // вычисляемое поле — сумма всех транзакций
}
