package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"vue-calc/internal/entity"
	"vue-calc/internal/usecase"
)

// CategoryHandler — HTTP-обработчик для работы с категориями расходов.
type CategoryHandler struct {
	uc *usecase.CategoryUseCase
}

// NewCategoryHandler — конструктор обработчика категорий.
func NewCategoryHandler(uc *usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

// Handle — обработка запросов к /api/categories и /api/categories/{id}.
func (h *CategoryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Требуется авторизация"}`, http.StatusUnauthorized)
		return
	}

	// Проверяем, есть ли ID в URL: /api/categories/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/categories")
	path = strings.TrimPrefix(path, "/")

	if path != "" && r.Method == http.MethodDelete {
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, `{"error": "Неверный ID категории"}`, http.StatusBadRequest)
			return
		}
		h.delete(w, id, userID)
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

// getAll — получить все категории пользователя.
func (h *CategoryHandler) getAll(w http.ResponseWriter, userID int) {
	categories, err := h.uc.GetAll(userID)
	if err != nil {
		http.Error(w, `{"error": "Ошибка получения категорий"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(categories)
}

// create — создать новую категорию.
func (h *CategoryHandler) create(w http.ResponseWriter, r *http.Request, userID int) {
	var category entity.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(category.Name) == "" {
		http.Error(w, `{"error": "Название категории обязательно"}`, http.StatusBadRequest)
		return
	}

	category.UserID = userID

	category, err := h.uc.Create(category)
	if err != nil {
		http.Error(w, `{"error": "Ошибка создания категории"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// delete — удалить категорию по ID.
func (h *CategoryHandler) delete(w http.ResponseWriter, id, userID int) {
	err := h.uc.Delete(id, userID)
	if err != nil {
		http.Error(w, `{"error": "Категория не найдена"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
