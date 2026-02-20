package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vue-calc/internal/usecase"
)

// StatisticsHandler — HTTP-обработчик для получения статистики.
type StatisticsHandler struct {
	uc *usecase.StatisticsUseCase
}

// NewStatisticsHandler — конструктор обработчика статистики.
func NewStatisticsHandler(uc *usecase.StatisticsUseCase) *StatisticsHandler {
	return &StatisticsHandler{uc: uc}
}

// Handle — обработка GET /api/statistics?from=...&to=...&account_id=...
func (h *StatisticsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Требуется авторизация"}`, http.StatusUnauthorized)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		http.Error(w, `{"error": "Параметры from и to обязательны"}`, http.StatusBadRequest)
		return
	}

	var accountID *int
	if aidStr := r.URL.Query().Get("account_id"); aidStr != "" {
		aid, err := strconv.Atoi(aidStr)
		if err != nil {
			http.Error(w, `{"error": "Неверный account_id"}`, http.StatusBadRequest)
			return
		}
		accountID = &aid
	}

	stats, err := h.uc.GetStatistics(userID, from, to, accountID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения статистики"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
