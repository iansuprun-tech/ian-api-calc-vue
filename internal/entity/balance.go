// Пакет entity содержит доменные структуры — основные объекты приложения.
// Эти структуры не зависят ни от БД, ни от HTTP — они описывают "что" есть в системе.
package entity

// Balance — баланс пользователя в определённой валюте.
// JSON-теги нужны для автоматической сериализации в HTTP-ответах.
type Balance struct {
	ID       int     `json:"id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
