package users_postgres_repository

import "github.com/horizoonn/todoapp/internal/core/domain"

type UserModel struct {
	ID          int
	Version     int
	FullName    string
	PhoneNumber *string
}

func userDomainFromModel(model UserModel) domain.User {
	return domain.NewUser(
		model.ID,
		model.Version,
		model.FullName,
		model.PhoneNumber,
	)
}

func userDomainsFromModels(userModels []UserModel) []domain.User {
	domains := make([]domain.User, len(userModels))

	for i, model := range userModels {
		domains[i] = userDomainFromModel(model)
	}

	return domains
}
