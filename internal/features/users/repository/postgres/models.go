package users_postgres_repository

import (
	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

type UserModel struct {
	ID          uuid.UUID
	Version     int
	FullName    string
	PhoneNumber *string
}

func (m *UserModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.Version,
		&m.FullName,
		&m.PhoneNumber,
	)
}

func modelToDomain(model UserModel) domain.User {
	return domain.NewUser(
		model.ID,
		model.Version,
		model.FullName,
		model.PhoneNumber,
	)
}

func modelsToDomains(models []UserModel) []domain.User {
	domains := make([]domain.User, len(models))

	for i, model := range models {
		domains[i] = modelToDomain(model)
	}

	return domains
}
