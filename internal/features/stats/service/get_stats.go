package stats_service

import (
	"context"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

func (s *StatsService) GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, error) {
	if filter.From != nil && filter.To != nil {
		if filter.To.Before(*filter.From) {
			return domain.Stats{}, fmt.Errorf("`to` must not be before `from`: %w", core_errors.ErrInvalidArgument)
		}
	}

	filter = normalizeDateRange(filter)

	cacheVersion := int64(0)
	cacheReadable := false
	if s.statsCache != nil {
		stats, found, version, err := s.statsCache.GetStats(ctx, filter)
		if err == nil {
			cacheVersion = version
			cacheReadable = true
		}
		if err == nil && found {
			return stats, nil
		}
	}

	stats, err := s.statsRepository.GetStats(ctx, filter)
	if err != nil {
		return domain.Stats{}, fmt.Errorf("get stats from repository: %w", err)
	}

	if s.statsCache != nil && cacheReadable {
		_ = s.statsCache.SetStats(ctx, filter, cacheVersion, stats)
	}

	return stats, nil
}

func normalizeDateRange(filter stats_feature.GetStatsFilter) stats_feature.GetStatsFilter {
	if filter.To == nil {
		return filter
	}

	toExclusive := filter.To.AddDate(0, 0, 1)
	filter.To = &toExclusive

	return filter
}
