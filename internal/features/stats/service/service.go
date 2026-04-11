package stats_service

import (
	"context"
	"time"

	"github.com/horizoonn/todoapp/internal/core/domain"
)

type StatsService struct {
	statsRepository StatsRepository
}

type StatsRepository interface {
	GetTasks(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error)
}

func NewStatsService(statsRepository StatsRepository) *StatsService {
	return &StatsService{
		statsRepository: statsRepository,
	}
}
