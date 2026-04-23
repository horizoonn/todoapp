package stats_service

import (
	"context"

	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

type StatsService struct {
	statsRepository StatsRepository
}

type StatsRepository interface {
	GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, error)
}

func NewStatsService(statsRepository StatsRepository) *StatsService {
	return &StatsService{
		statsRepository: statsRepository,
	}
}
