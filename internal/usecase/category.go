package usecase

import "vue-calc/internal/entity"

// CategoryRepository — интерфейс репозитория категорий.
type CategoryRepository interface {
	GetAllByUserID(userID int) ([]entity.Category, error)
	Create(category entity.Category) (entity.Category, error)
	Delete(id, userID int) error
}

// CategoryUseCase — бизнес-логика для работы с категориями расходов.
type CategoryUseCase struct {
	repo CategoryRepository
}

// NewCategoryUseCase — конструктор юзкейса категорий.
func NewCategoryUseCase(repo CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

// GetAll — получить все категории пользователя.
func (uc *CategoryUseCase) GetAll(userID int) ([]entity.Category, error) {
	return uc.repo.GetAllByUserID(userID)
}

// Create — создать новую категорию.
func (uc *CategoryUseCase) Create(category entity.Category) (entity.Category, error) {
	return uc.repo.Create(category)
}

// Delete — удалить категорию.
func (uc *CategoryUseCase) Delete(id, userID int) error {
	return uc.repo.Delete(id, userID)
}
