package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
)

func (s *TasksService) PatchTask(ctx context.Context, id uuid.UUID, patch domain.TaskPatch) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task from repository: %w", err)
	}

	if err := task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf("apply task patch: %w", err)
	}

	patchedTask, err := s.tasksRepository.UpdateTask(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task in repository: %w", err)
	}

	return patchedTask, nil
}
