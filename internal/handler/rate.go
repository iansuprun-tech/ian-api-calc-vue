package handler

import (
	"encoding/json"
	"net/http"

	"vue-calc/internal/usecase"
)

// RateHandler — HTTP-обработчик для эндпоинта /api/rates.
type RateHandler struct {
	uc *usecase.RateUseCase
}

// NewRateHandler — конструктор.
func NewRateHandler(uc *usecase.RateUseCase) *RateHandler {
	return &RateHandler{uc: uc}
}

// Handle обрабатывает GET /api/rates — возвращает список всех курсов валют.
func (h *RateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	rates, err := h.uc.GetAll()
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения курсов"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rates)
}
