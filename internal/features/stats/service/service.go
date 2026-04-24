package stats_service

import (
	"context"

	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

type StatsService struct {
	statsRepository StatsRepository
	statsCache      StatsCache
}

type StatsRepository interface {
	GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, error)
}

type StatsCache interface {
	GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, bool, int64, error)
	SetStats(ctx context.Context, filter stats_feature.GetStatsFilter, version int64, stats domain.Stats) error
}

func NewStatsService(statsRepository StatsRepository, statsCache StatsCache) *StatsService {
	return &StatsService{
		statsRepository: statsRepository,
		statsCache:      statsCache,
	}
}
