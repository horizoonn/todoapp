package stats_redis_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
)

type StatsRepository struct {
	pool core_redis_pool.Pool
}

func NewStatsRepository(pool core_redis_pool.Pool) *StatsRepository {
	return &StatsRepository{
		pool: pool,
	}
}

func (r *StatsRepository) InvalidateStats(ctx context.Context, userID uuid.UUID) error {
	if err := r.pool.Incr(ctx, statsVersionAllKey()).Err(); err != nil {
		return fmt.Errorf("increment next global stats version: %w", err)
	}

	if err := r.pool.Incr(ctx, statsVersionUserKey(userID)).Err(); err != nil {
		return fmt.Errorf("increment next stats version: %w", err)
	}

	return nil
}
