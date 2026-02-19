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
type AccountHandler struct {
	uc *usecase.AccountUseCase
}

// NewAccountHandler — конструктор обработчика счетов.
func NewAccountHandler(uc *usecase.AccountUseCase) *AccountHandler {
	return &AccountHandler{uc: uc}
}

// HandleList — обработка запросов к /api/accounts (список + создание).
func (h *AccountHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Требуется авторизация"}`, http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAll(w, userID)
	case http.MethodPost:
		h.create(w, r, userID)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// HandleByID — обработка запросов к /api/accounts/{id}.
func (h *AccountHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Требуется авторизация"}`, http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Неверный ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getByID(w, id, userID)
	case http.MethodPut:
		h.updateComment(w, r, id, userID)
	case http.MethodDelete:
		h.delete(w, id, userID)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// getAll — получить все счета пользователя.
func (h *AccountHandler) getAll(w http.ResponseWriter, userID int) {
	accounts, err := h.uc.GetAll(userID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения счетов"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

// create — создать новый счёт.
func (h *AccountHandler) create(w http.ResponseWriter, r *http.Request, userID int) {
	var account entity.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	account.UserID = userID

	account, err := h.uc.Create(account)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания счёта"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// getByID — получить один счёт по ID.
func (h *AccountHandler) getByID(w http.ResponseWriter, id, userID int) {
	account, err := h.uc.GetByID(id, userID)
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

// updateComment — обновить комментарий счёта.
func (h *AccountHandler) updateComment(w http.ResponseWriter, r *http.Request, id, userID int) {
	var body struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if err := h.uc.UpdateComment(id, userID, body.Comment); err != nil {
		http.Error(w, `{"error": "Счёт не найден"}`, http.StatusNotFound)
		return
	}

	account, err := h.uc.GetByID(id, userID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения счёта"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(account)
}

// delete — удалить счёт по ID.
func (h *AccountHandler) delete(w http.ResponseWriter, id, userID int) {
	rowsAffected, err := h.uc.Delete(id, userID)
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
