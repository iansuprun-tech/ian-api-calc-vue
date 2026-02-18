// Пакет usecase содержит бизнес-логику приложения.
// Он определяет ЧТО делает приложение, но не КАК хранятся данные.
// Для работы с данными используются интерфейсы (контракты),
// а конкретные реализации подставляются снаружи (Dependency Injection).
package usecase

import "vue-calc/internal/entity"

// BalanceRepository — интерфейс (контракт) для работы с балансами в БД.
// Usecase не знает, какая БД используется — PostgreSQL, MySQL или что-то ещё.
// Главное — чтобы реализация выполняла эти методы.
type BalanceRepository interface {
	GetAll() ([]entity.Balance, error)
	GetByID(id int) (entity.Balance, error)
	Create(balance entity.Balance) (entity.Balance, error)
	Update(id int, balance entity.Balance) error
	Delete(id int) (int64, error)
	Exists(id int) (bool, error)
}

// BalanceUseCase — бизнес-логика для работы с балансами.
// Принимает BalanceRepository через конструктор — это и есть Dependency Injection.
type BalanceUseCase struct {
	repo BalanceRepository
}

// NewBalanceUseCase — конструктор. Принимает любую реализацию BalanceRepository.
func NewBalanceUseCase(repo BalanceRepository) *BalanceUseCase {
	return &BalanceUseCase{repo: repo}
}

// GetAll возвращает все балансы.
func (uc *BalanceUseCase) GetAll() ([]entity.Balance, error) {
	return uc.repo.GetAll()
}

// GetByID возвращает баланс по ID.
func (uc *BalanceUseCase) GetByID(id int) (entity.Balance, error) {
	return uc.repo.GetByID(id)
}

// Create создаёт новый баланс и возвращает его с присвоенным ID.
func (uc *BalanceUseCase) Create(balance entity.Balance) (entity.Balance, error) {
	return uc.repo.Create(balance)
}

// Update обновляет баланс по ID. Сначала проверяет, что запись существует.
func (uc *BalanceUseCase) Update(id int, balance entity.Balance) error {
	return uc.repo.Update(id, balance)
}

// Exists проверяет, существует ли баланс с данным ID.
func (uc *BalanceUseCase) Exists(id int) (bool, error) {
	return uc.repo.Exists(id)
}

// Delete удаляет баланс по ID. Возвращает количество удалённых строк.
func (uc *BalanceUseCase) Delete(id int) (int64, error) {
	return uc.repo.Delete(id)
}
