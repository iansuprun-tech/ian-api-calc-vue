package postgres

import (
	"database/sql"
	"vue-calc/internal/entity"
)

// CategoryRepo — репозиторий для работы с категориями в PostgreSQL.
type CategoryRepo struct {
	db *sql.DB
}

// NewCategoryRepo — конструктор репозитория категорий.
func NewCategoryRepo(db *sql.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

// GetAllByUserID — получить все категории пользователя.
func (r *CategoryRepo) GetAllByUserID(userID int) ([]entity.Category, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, name, created_at FROM categories WHERE user_id = $1 AND deleted_at IS NULL ORDER BY name",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []entity.Category{}
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// Create — создать новую категорию.
func (r *CategoryRepo) Create(category entity.Category) (entity.Category, error) {
	err := r.db.QueryRow(
		"INSERT INTO categories (user_id, name) VALUES ($1, $2) RETURNING id, created_at",
		category.UserID, category.Name,
	).Scan(&category.ID, &category.CreatedAt)
	return category, err
}

// Delete — мягко удалить категорию по ID (только если принадлежит пользователю).
func (r *CategoryRepo) Delete(id, userID int) error {
	res, err := r.db.Exec(
		"UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL",
		id, userID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
