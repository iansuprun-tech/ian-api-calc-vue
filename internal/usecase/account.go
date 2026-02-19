package usecase

import "vue-calc/internal/entity"

// AccountRepository — интерфейс репозитория счетов.
// Все операции фильтруются по user_id.
type AccountRepository interface {
	GetAll(userID int) ([]entity.Account, error)
	GetByID(id, userID int) (entity.Account, error)
	Create(account entity.Account) (entity.Account, error)
	Delete(id, userID int) (int64, error)
	Exists(id, userID int) (bool, error)
}

// AccountUseCase — бизнес-логика для работы со счетами.
type AccountUseCase struct {
	repo AccountRepository
}

// NewAccountUseCase — конструктор юзкейса счетов.
func NewAccountUseCase(repo AccountRepository) *AccountUseCase {
	return &AccountUseCase{repo: repo}
}

// GetAll — получить все счета пользователя.
func (uc *AccountUseCase) GetAll(userID int) ([]entity.Account, error) {
	return uc.repo.GetAll(userID)
}

// GetByID — получить счёт по ID (с проверкой принадлежности пользователю).
func (uc *AccountUseCase) GetByID(id, userID int) (entity.Account, error) {
	return uc.repo.GetByID(id, userID)
}

// Create — создать новый счёт.
func (uc *AccountUseCase) Create(account entity.Account) (entity.Account, error) {
	return uc.repo.Create(account)
}

// Delete — удалить счёт (транзакции удалятся каскадом).
func (uc *AccountUseCase) Delete(id, userID int) (int64, error) {
	return uc.repo.Delete(id, userID)
}

// Exists — проверить существование счёта у пользователя.
func (uc *AccountUseCase) Exists(id, userID int) (bool, error) {
	return uc.repo.Exists(id, userID)
}
