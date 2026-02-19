package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"vue-calc/internal/entity"
	"vue-calc/internal/usecase"
)

// TransactionHandler — HTTP-обработчик для работы с операциями (транзакциями) по счетам.
type TransactionHandler struct {
	txUC      *usecase.TransactionUseCase
	accountUC *usecase.AccountUseCase
}

// NewTransactionHandler — конструктор обработчика транзакций.
// Принимает юзкейс транзакций и юзкейс счетов (для проверки существования счёта).
func NewTransactionHandler(txUC *usecase.TransactionUseCase, accountUC *usecase.AccountUseCase) *TransactionHandler {
	return &TransactionHandler{txUC: txUC, accountUC: accountUC}
}

// Handle — обработка запросов к /api/accounts/{id}/transactions.
// GET  — получить историю операций по счёту (новые сверху).
// POST — добавить операцию {amount, comment}.
func (h *TransactionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Извлекаем account_id из URL: /api/accounts/123/transactions → "123"
	path := strings.TrimPrefix(r.URL.Path, "/api/accounts/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 || parts[1] != "transactions" {
		http.Error(w, `{"error": "Неверный URL"}`, http.StatusBadRequest)
		return
	}

	accountID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, `{"error": "Неверный ID счёта"}`, http.StatusBadRequest)
		return
	}

	// Проверяем, что счёт существует
	exists, err := h.accountUC.Exists(accountID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка проверки счёта"}`, http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, `{"error": "Счёт не найден"}`, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getByAccountID(w, accountID)
	case http.MethodPost:
		h.create(w, r, accountID)
	default:
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
	}
}

// getByAccountID — получить все транзакции по счёту.
func (h *TransactionHandler) getByAccountID(w http.ResponseWriter, accountID int) {
	transactions, err := h.txUC.GetByAccountID(accountID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения операций"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(transactions)
}

// create — создать новую транзакцию по счёту.
func (h *TransactionHandler) create(w http.ResponseWriter, r *http.Request, accountID int) {
	var transaction entity.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	// Привязываем транзакцию к счёту из URL
	transaction.AccountID = accountID

	transaction, err := h.txUC.Create(transaction)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания операции"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}
