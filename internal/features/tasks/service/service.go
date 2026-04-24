package tasks_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

type TasksService struct {
	tasksRepository       TasksRepository
	tasksCache            TasksCache
	statsCacheInvalidator StatsCacheInvalidator
}

type TasksRepository interface {
	SaveTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter) ([]domain.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (domain.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error)
}

type TasksCache interface {
	GetTask(ctx context.Context, id uuid.UUID) (domain.Task, bool, error)
	SetTask(ctx context.Context, task domain.Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
	GetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter) ([]domain.Task, bool, int64, error)
	SetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter, version int64, tasks []domain.Task) error
	InvalidateTasks(ctx context.Context, userID uuid.UUID) error
}

type StatsCacheInvalidator interface {
	InvalidateStats(ctx context.Context, userID uuid.UUID) error
}

func NewTasksService(tasksRepository TasksRepository, tasksCache TasksCache, statsCacheInvalidator StatsCacheInvalidator) *TasksService {
	return &TasksService{
		tasksRepository:       tasksRepository,
		tasksCache:            tasksCache,
		statsCacheInvalidator: statsCacheInvalidator,
	}
}
