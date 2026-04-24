package tasks_redis_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_redis_pool "github.com/horizoonn/todoapp/internal/core/repository/redis/pool"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

func (r *TasksRepository) GetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter) ([]domain.Task, bool, int64, error) {
	version, err := r.getTasksListVersion(ctx, filter)
	if err != nil {
		return nil, false, 0, fmt.Errorf("get tasks list version: %w", err)
	}

	data, err := r.pool.Get(ctx, tasksListKey(filter, version)).Bytes()
	if err != nil {
		if errors.Is(err, core_redis_pool.NotFound) {
			return nil, false, version, nil
		}

		return nil, false, 0, fmt.Errorf("get tasks from redis: %w", err)
	}

	var taskModels []TaskModel
	if err := json.Unmarshal(data, &taskModels); err != nil {
		return nil, false, 0, fmt.Errorf("unmarshal tasks from redis: %w", err)
	}

	return modelsToDomains(taskModels), true, version, nil
}

func (r *TasksRepository) getTasksListVersion(ctx context.Context, filter tasks_feature.GetTasksFilter) (int64, error) {
	key := tasksListVersionAllKey()
	if filter.UserID != nil {
		key = tasksListVersionUserKey(*filter.UserID)
	}

	data, err := r.pool.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, core_redis_pool.NotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("get tasks list version from redis: %w", err)
	}

	version, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse tasks list version: %w", err)
	}

	return version, nil
}
