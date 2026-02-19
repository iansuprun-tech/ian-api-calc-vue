package usecase

import (
	"log"
	"time"

	"vue-calc/internal/entity"
)

// RateRepository — интерфейс для работы с курсами валют в БД.
type RateRepository interface {
	GetAll() ([]entity.Rate, error)
	Upsert(currency string, rateToUSD float64) error
}

// RateFetcher — интерфейс для получения курсов из внешнего API.
// Отделяем HTTP-клиент от бизнес-логики, чтобы usecase не зависел от конкретного API.
type RateFetcher interface {
	FetchRates() (*entity.ExchangeRateResponse, error)
}

// RateUseCase — бизнес-логика для работы с курсами валют.
type RateUseCase struct {
	repo    RateRepository
	fetcher RateFetcher
}

// NewRateUseCase — конструктор.
func NewRateUseCase(repo RateRepository, fetcher RateFetcher) *RateUseCase {
	return &RateUseCase{repo: repo, fetcher: fetcher}
}

// GetAll возвращает все курсы валют из БД.
func (uc *RateUseCase) GetAll() ([]entity.Rate, error) {
	return uc.repo.GetAll()
}

// SaveRates получает курсы из внешнего API и сохраняет их в БД.
// Для каждой валюты вычисляет курс к USD (1 / rate) и делает upsert.
func (uc *RateUseCase) SaveRates() {
	rateResponse, err := uc.fetcher.FetchRates()
	if err != nil {
		log.Println("Ошибка получения курсов:", err)
		return
	}

	if rateResponse.Result != "success" {
		log.Println("API вернул ошибку, result:", rateResponse.Result)
		return
	}

	for currency, rate := range rateResponse.ConversionRates {
		var rateToUSD float64
		if rate > 0 {
			rateToUSD = 1.0 / rate
		}

		if err := uc.repo.Upsert(currency, rateToUSD); err != nil {
			log.Println("Ошибка сохранения курса для", currency, ":", err)
		}
	}

	log.Println("Курсы валют обновлены успешно!")
}

// StartUpdater запускает фоновое обновление курсов каждый час.
// Первый запрос выполняется сразу, далее — по таймеру.
func (uc *RateUseCase) StartUpdater() {
	uc.SaveRates()
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for range ticker.C {
			uc.SaveRates()
		}
	}()
}
