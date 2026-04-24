package tasks_redis_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *TasksRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if err := r.pool.Del(ctx, taskKey(id)).Err(); err != nil {
		return fmt.Errorf("delete task from redis: %w", err)
	}

	return nil
}
