package usecase

import "vue-calc/internal/entity"

// StatisticsRepository — интерфейс репозитория статистики.
type StatisticsRepository interface {
	GetStatistics(userID int, from, to string, accountID *int) (entity.StatisticsResponse, error)
}

// StatisticsUseCase — бизнес-логика для получения статистики.
type StatisticsUseCase struct {
	repo StatisticsRepository
}

// NewStatisticsUseCase — конструктор юзкейса статистики.
func NewStatisticsUseCase(repo StatisticsRepository) *StatisticsUseCase {
	return &StatisticsUseCase{repo: repo}
}

// GetStatistics — получить агрегированную статистику за период.
func (uc *StatisticsUseCase) GetStatistics(userID int, from, to string, accountID *int) (entity.StatisticsResponse, error) {
	return uc.repo.GetStatistics(userID, from, to, accountID)
}
