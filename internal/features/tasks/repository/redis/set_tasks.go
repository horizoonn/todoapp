package tasks_redis_repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

func (r *TasksRepository) SetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter, version int64, tasks []domain.Task) error {
	taskModels := modelsFromDomains(tasks)

	data, err := json.Marshal(taskModels)
	if err != nil {
		return fmt.Errorf("marshal tasks for redis: %w", err)
	}

	if err := r.pool.Set(ctx, tasksListKey(filter, version), data, r.pool.TTL()).Err(); err != nil {
		return fmt.Errorf("set tasks in redis: %w", err)
	}

	return nil
}
