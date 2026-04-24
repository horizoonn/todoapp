package tasks_redis_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
)

func (r *TasksRepository) GetTask(ctx context.Context, id uuid.UUID) (domain.Task, bool, error) {
	data, err := r.pool.Get(ctx, taskKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, core_redis_pool.NotFound) {
			return domain.Task{}, false, nil
		}

		return domain.Task{}, false, fmt.Errorf("get task from redis: %w", err)
	}

	var taskModel TaskModel
	if err := json.Unmarshal(data, &taskModel); err != nil {
		return domain.Task{}, false, fmt.Errorf("unmarshal task from redis: %w", err)
	}

	return modelToDomain(taskModel), true, nil
}
