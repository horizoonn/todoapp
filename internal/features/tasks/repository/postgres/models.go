package tasks_postgres_repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

type TaskModel struct {
	ID      uuid.UUID
	Version int

	Title       string
	Description *string
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time

	AuthorUserID uuid.UUID
}

func (m *TaskModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.Version,
		&m.Title,
		&m.Description,
		&m.Completed,
		&m.CreatedAt,
		&m.CompletedAt,
		&m.AuthorUserID,
	)
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

func modelsToDomains(taskModels []TaskModel) []domain.Task {
	domains := make([]domain.Task, len(taskModels))

	for i, model := range taskModels {
		domains[i] = modelToDomain(model)
	}

	return domains
}
