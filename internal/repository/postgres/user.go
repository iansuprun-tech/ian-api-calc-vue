package postgres

import (
	"database/sql"
	"vue-calc/internal/entity"
)

// UserRepo — репозиторий для работы с пользователями в PostgreSQL.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo — конструктор репозитория пользователей.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create — создать нового пользователя. Возвращает созданного пользователя с ID.
func (r *UserRepo) Create(email, passwordHash string) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at",
		email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	return user, err
}

// GetByEmail — найти пользователя по email.
func (r *UserRepo) GetByEmail(email string) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	return user, err
}
