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

// AccountHandler — HTTP-обработчик для работы со счетами.
// Обрабатывает запросы на создание, получение и удаление счетов.
type AccountHandler struct {
	uc *usecase.AccountUseCase
}

// NewAccountHandler — конструктор обработчика счетов.
func NewAccountHandler(uc *usecase.AccountUseCase) *AccountHandler {
	return &AccountHandler{uc: uc}
}

// HandleList — обработка запросов к /api/accounts (список + создание).
// GET  — получить все счета с балансами.
// POST — создать новый счёт {currency, comment}.
func (h *AccountHandler) HandleList(w http.ResponseWriter, r *http.Request) {
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

// HandleByID — обработка запросов к /api/accounts/{id}.
// GET    — получить один счёт с балансом.
// DELETE — удалить счёт (транзакции удалятся каскадом).
func (h *AccountHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Извлекаем ID из URL: /api/accounts/123 → "123"
	idStr := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Неверный ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getByID(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// getAll — получить все счета с вычисленными балансами.
func (h *AccountHandler) getAll(w http.ResponseWriter, _ *http.Request) {
	accounts, err := h.uc.GetAll()
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения счетов"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

// create — создать новый счёт.
func (h *AccountHandler) create(w http.ResponseWriter, r *http.Request) {
	var account entity.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	account, err := h.uc.Create(account)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания счёта"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// getByID — получить один счёт по ID.
func (h *AccountHandler) getByID(w http.ResponseWriter, _ *http.Request, id int) {
	account, err := h.uc.GetByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Счёт не найден"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения счёта"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(account)
}

// delete — удалить счёт по ID.
func (h *AccountHandler) delete(w http.ResponseWriter, _ *http.Request, id int) {
	rowsAffected, err := h.uc.Delete(id)
	if err != nil {
		http.Error(w, `{"error": "Ошибка удаления счёта"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "Счёт не найден"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
