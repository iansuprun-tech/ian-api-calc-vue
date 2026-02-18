// Пакет handler содержит HTTP-обработчики.
// Обработчик принимает HTTP-запрос, вызывает usecase и формирует ответ.
// Он не содержит бизнес-логики и не знает про БД.
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"vue-calc/internal/entity"
	"vue-calc/internal/usecase"
)

// BalanceHandler — HTTP-обработчики для эндпоинтов /api/balances.
type BalanceHandler struct {
	uc *usecase.BalanceUseCase
}

// NewBalanceHandler — конструктор.
func NewBalanceHandler(uc *usecase.BalanceUseCase) *BalanceHandler {
	return &BalanceHandler{uc: uc}
}

// HandleList обрабатывает /api/balances (GET — список, POST — создание).
func (h *BalanceHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getAll(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// HandleByID обрабатывает /api/balances/{id} (GET, PUT, DELETE).
func (h *BalanceHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Извлекаем ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/balances/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Неверный ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getByID(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// getAll возвращает список всех балансов.
func (h *BalanceHandler) getAll(w http.ResponseWriter, _ *http.Request) {
	balances, err := h.uc.GetAll()
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения балансов"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(balances)
}

// create создаёт новый баланс из тела запроса.
func (h *BalanceHandler) create(w http.ResponseWriter, r *http.Request) {
	var balance entity.Balance
	if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	balance, err := h.uc.Create(balance)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания баланса"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(balance)
}

// getByID возвращает баланс по ID.
func (h *BalanceHandler) getByID(w http.ResponseWriter, _ *http.Request, id int) {
	balance, err := h.uc.GetByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Баланс не найден"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения баланса"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(balance)
}

// update обновляет баланс по ID.
func (h *BalanceHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var balance entity.Balance
	if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
		http.Error(w, `{"error": "неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем существование
	exists, err := h.uc.Exists(id)
	if err != nil || !exists {
		http.Error(w, `{"error": "Баланс не найден"}`, http.StatusNotFound)
		return
	}

	if err := h.uc.Update(id, balance); err != nil {
		http.Error(w, `{"error": "Ошибка обновления баланса"}`, http.StatusInternalServerError)
		return
	}

	balance.ID = id
	json.NewEncoder(w).Encode(balance)
}

// delete удаляет баланс по ID.
func (h *BalanceHandler) delete(w http.ResponseWriter, _ *http.Request, id int) {
	rowsAffected, err := h.uc.Delete(id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка удаления баланса"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "Баланс не найден"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
