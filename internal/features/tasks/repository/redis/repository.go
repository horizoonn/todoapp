package tasks_redis_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
)

type TasksRepository struct {
	pool core_redis_pool.Pool
}

func NewTasksRepository(pool core_redis_pool.Pool) *TasksRepository {
	return &TasksRepository{
		pool: pool,
	}
}

func (r *TasksRepository) InvalidateTasks(ctx context.Context, userID uuid.UUID) error {
	if err := r.pool.Incr(ctx, tasksListVersionAllKey()).Err(); err != nil {
		return fmt.Errorf("increment global tasks list version: %w", err)
	}

	if err := r.pool.Incr(ctx, tasksListVersionUserKey(userID)).Err(); err != nil {
		return fmt.Errorf("increment user tasks list version: %w", err)
	}

	return nil
}
