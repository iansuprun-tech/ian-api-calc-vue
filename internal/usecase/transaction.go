package usecase

import "vue-calc/internal/entity"

// TransactionRepository — интерфейс репозитория транзакций.
// Определяет контракт для слоя данных.
type TransactionRepository interface {
	GetByAccountID(accountID int) ([]entity.Transaction, error)
	Create(transaction entity.Transaction) (entity.Transaction, error)
	Delete(id, accountID int) error
	Update(id, accountID int, transaction entity.Transaction) (entity.Transaction, error)
}

// TransactionUseCase — бизнес-логика для работы с транзакциями (операциями по счетам).
type TransactionUseCase struct {
	repo TransactionRepository
}

// NewTransactionUseCase — конструктор юзкейса транзакций.
func NewTransactionUseCase(repo TransactionRepository) *TransactionUseCase {
	return &TransactionUseCase{repo: repo}
}

// GetByAccountID — получить все транзакции по счёту (новые сверху).
func (uc *TransactionUseCase) GetByAccountID(accountID int) ([]entity.Transaction, error) {
	return uc.repo.GetByAccountID(accountID)
}

// Create — создать новую транзакцию (пополнение или списание).
func (uc *TransactionUseCase) Create(transaction entity.Transaction) (entity.Transaction, error) {
	return uc.repo.Create(transaction)
}

// Delete — удалить транзакцию по ID.
func (uc *TransactionUseCase) Delete(id, accountID int) error {
	return uc.repo.Delete(id, accountID)
}

// Update — обновить транзакцию по ID.
func (uc *TransactionUseCase) Update(id, accountID int, transaction entity.Transaction) (entity.Transaction, error) {
	return uc.repo.Update(id, accountID, transaction)
}
