package usecase

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"vue-calc/internal/entity"
)

// UserRepository — интерфейс репозитория пользователей.
type UserRepository interface {
	Create(email, passwordHash string) (entity.User, error)
	GetByEmail(email string) (entity.User, error)
}

// AuthUseCase — бизнес-логика аутентификации.
type AuthUseCase struct {
	repo UserRepository
}

// NewAuthUseCase — конструктор.
func NewAuthUseCase(repo UserRepository) *AuthUseCase {
	return &AuthUseCase{repo: repo}
}

// Register — регистрация нового пользователя.
func (uc *AuthUseCase) Register(email, password string) (entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err
	}
	return uc.repo.Create(email, string(hash))
}

// Login — вход пользователя, возвращает JWT-токен.
func (uc *AuthUseCase) Login(email, password string) (string, error) {
	user, err := uc.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("неверный email или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("неверный email или пароль")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "vue-calc-default-secret"
	}

	return token.SignedString([]byte(secret))
}
