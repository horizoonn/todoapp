package tasks_redis_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
)

func (r *TasksRepository) SetTask(ctx context.Context, task domain.Task) error {
	taskModel := modelFromDomain(task)

	data, err := json.Marshal(taskModel)
	if err != nil {
		return fmt.Errorf("marshal task for redis: %w", err)
	}

	if err := r.pool.Set(ctx, taskKey(task.ID), data, r.pool.TTL()).Err(); err != nil {
		return fmt.Errorf("set task in redis: %w", err)
	}

	return nil
}
