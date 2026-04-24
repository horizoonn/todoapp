package stats_redis_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

func (r *StatsRepository) SetStats(ctx context.Context, filter stats_feature.GetStatsFilter, version int64, stats domain.Stats) error {
	key := statsKey(filter, version)

	statsModel := modelFromDomain(stats)

	data, err := json.Marshal(statsModel)
	if err != nil {
		return fmt.Errorf("marshal stats for redis: %w", err)
	}

	if err := r.pool.Set(ctx, key, data, r.pool.TTL()).Err(); err != nil {
		return fmt.Errorf("set stats in redis: %w", err)
	}

	return nil
}
