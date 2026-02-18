package entity

// Rate — курс валюты к USD, хранящийся в базе данных.
type Rate struct {
	ID        int     `json:"id"`
	Currency  string  `json:"currency"`
	RateToUSD float64 `json:"rate_to_usd"`
	UpdatedAt string  `json:"updated_at"`
}

// ExchangeRateResponse — структура ответа от внешнего API exchangerate-api.com.
// ConversionRates содержит маппинг "валюта -> курс относительно базовой валюты".
type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}
