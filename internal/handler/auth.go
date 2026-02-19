package handler

import (
	"encoding/json"
	"net/http"

	"vue-calc/internal/usecase"
)

// AuthHandler — HTTP-обработчик для регистрации и входа.
type AuthHandler struct {
	uc *usecase.AuthUseCase
}

// NewAuthHandler — конструктор.
func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// authRequest — тело запроса на регистрацию/вход.
type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleRegister — POST /api/register.
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email и пароль обязательны"}`, http.StatusBadRequest)
		return
	}

	user, err := h.uc.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error": "Пользователь с таким email уже существует"}`, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// HandleLogin — POST /api/login.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Неверный формат JSON"}`, http.StatusBadRequest)
		return
	}

	token, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error": "Неверный email или пароль"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
