package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	"github.com/horizoonn/todoapp/internal/core/pagination"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

const (
	defaultTasksLimit  = 100
	maxTasksLimit      = 100
	defaultTasksOffset = 0
)

func (s *TasksService) GetTasks(ctx context.Context, userID *uuid.UUID, limit *int, offset *int) ([]domain.Task, error) {
	page, err := pagination.Normalize(limit, offset, defaultTasksLimit, maxTasksLimit, defaultTasksOffset)
	if err != nil {
		return nil, fmt.Errorf("normalize pagination: %w", err)
	}

	filter := tasks_feature.NewGetTasksFilter(userID, page.Limit, page.Offset)

	cacheVersion := int64(0)
	cacheReadable := false
	if s.tasksCache != nil {
		tasks, found, version, err := s.tasksCache.GetTasks(ctx, filter)
		if err == nil {
			cacheVersion = version
			cacheReadable = true
		}
		if err == nil && found {
			return tasks, nil
		}
	}

	tasks, err := s.tasksRepository.GetTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get tasks from repository: %w", err)
	}

	if s.tasksCache != nil && cacheReadable {
		_ = s.tasksCache.SetTasks(ctx, filter, cacheVersion, tasks)
	}

	return tasks, nil
}
