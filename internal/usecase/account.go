package usecase

import "vue-calc/internal/entity"

// AccountRepository — интерфейс репозитория счетов.
// Определяет контракт для слоя данных, позволяя подменять реализацию (например, для тестов).
type AccountRepository interface {
	GetAll() ([]entity.Account, error)
	GetByID(id int) (entity.Account, error)
	Create(account entity.Account) (entity.Account, error)
	Delete(id int) (int64, error)
	Exists(id int) (bool, error)
}

// AccountUseCase — бизнес-логика для работы со счетами.
type AccountUseCase struct {
	repo AccountRepository
}

// NewAccountUseCase — конструктор юзкейса счетов.
func NewAccountUseCase(repo AccountRepository) *AccountUseCase {
	return &AccountUseCase{repo: repo}
}

// GetAll — получить все счета с балансами.
func (uc *AccountUseCase) GetAll() ([]entity.Account, error) {
	return uc.repo.GetAll()
}

// GetByID — получить счёт по ID.
func (uc *AccountUseCase) GetByID(id int) (entity.Account, error) {
	return uc.repo.GetByID(id)
}

// Create — создать новый счёт.
func (uc *AccountUseCase) Create(account entity.Account) (entity.Account, error) {
	return uc.repo.Create(account)
}

// Delete — удалить счёт (транзакции удалятся каскадом).
func (uc *AccountUseCase) Delete(id int) (int64, error) {
	return uc.repo.Delete(id)
}

// Exists — проверить существование счёта.
func (uc *AccountUseCase) Exists(id int) (bool, error) {
	return uc.repo.Exists(id)
}
