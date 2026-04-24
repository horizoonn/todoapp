package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *TasksService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("get task from repository: %w", err)
	}

	if err := s.tasksRepository.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("delete task from repository: %w", err)
	}

	if s.tasksCache != nil {
		_ = s.tasksCache.DeleteTask(ctx, id)
		_ = s.tasksCache.InvalidateTasks(ctx, task.AuthorUserID)
	}

	if s.statsCacheInvalidator != nil {
		_ = s.statsCacheInvalidator.InvalidateStats(ctx, task.AuthorUserID)
	}

	return nil
}
