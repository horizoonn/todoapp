package tasks_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

type TasksService struct {
	tasksRepository TasksRepository
}

type TasksRepository interface {
	SaveTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter) ([]domain.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (domain.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error)
}

func NewTasksService(tasksRepository TasksRepository) *TasksService {
	return &TasksService{
		tasksRepository: tasksRepository,
	}
}
