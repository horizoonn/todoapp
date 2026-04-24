package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
)

func (s *TasksService) GetTask(ctx context.Context, id uuid.UUID) (domain.Task, error) {
	if s.tasksCache != nil {
		task, found, err := s.tasksCache.GetTask(ctx, id)
		if err == nil && found {
			return task, nil
		}
	}

	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task from repository: %w", err)
	}

	if s.tasksCache != nil {
		_ = s.tasksCache.SetTask(ctx, task)
	}

	return task, nil
}
