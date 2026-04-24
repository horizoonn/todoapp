package stats_redis_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

func (r *StatsRepository) GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, bool, int64, error) {
	version, err := r.getStatsVersion(ctx, filter)
	if err != nil {
		return domain.Stats{}, false, 0, fmt.Errorf("get stats version: %w", err)
	}

	key := statsKey(filter, version)

	data, err := r.pool.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, core_redis_pool.NotFound) {
			return domain.Stats{}, false, version, nil
		}

		return domain.Stats{}, false, 0, fmt.Errorf("get stats from redis: %w", err)
	}

	var statsModel StatsModel
	if err := json.Unmarshal(data, &statsModel); err != nil {
		return domain.Stats{}, false, 0, fmt.Errorf("unmarshal stats from redis: %w", err)
	}

	return modelToDomain(statsModel), true, version, nil
}

func (r *StatsRepository) getStatsVersion(ctx context.Context, filter stats_feature.GetStatsFilter) (int64, error) {
	key := statsVersionAllKey()
	if filter.UserID != nil {
		key = statsVersionUserKey(*filter.UserID)
	}

	data, err := r.pool.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, core_redis_pool.NotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("get stats version from redis: %w", err)
	}

	version, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse stats version: %w", err)
	}

	return version, nil
}
