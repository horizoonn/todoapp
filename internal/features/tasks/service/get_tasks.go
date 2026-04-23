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

	tasks, err := s.tasksRepository.GetTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get tasks from repository: %w", err)
	}

	return tasks, nil
}
