package tasks_redis_repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
)

type TaskModel struct {
	ID      uuid.UUID `json:"id"`
	Version int       `json:"version"`

	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`

	AuthorUserID uuid.UUID `json:"author_user_id"`
}

func modelToDomain(model TaskModel) domain.Task {
	return domain.NewTask(
		model.ID,
		model.Version,
		model.Title,
		model.Description,
		model.Completed,
		model.CreatedAt,
		model.CompletedAt,
		model.AuthorUserID,
	)
}

func modelFromDomain(task domain.Task) TaskModel {
	return TaskModel{
		ID:           task.ID,
		Version:      task.Version,
		Title:        task.Title,
		Description:  task.Description,
		Completed:    task.Completed,
		CreatedAt:    task.CreatedAt,
		CompletedAt:  task.CompletedAt,
		AuthorUserID: task.AuthorUserID,
	}
}

func modelsToDomains(models []TaskModel) []domain.Task {
	domains := make([]domain.Task, len(models))

	for i, model := range models {
		domains[i] = modelToDomain(model)
	}

	return domains
}

func modelsFromDomains(tasks []domain.Task) []TaskModel {
	models := make([]TaskModel, len(tasks))

	for i, task := range tasks {
		models[i] = modelFromDomain(task)
	}

	return models
}
