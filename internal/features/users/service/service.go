package users_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	users_feature "github.com/horizoonn/todoapp/internal/features/users"
)

type UsersService struct {
	usersRepository UsersRepository
}

type UsersRepository interface {
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUsers(ctx context.Context, filter users_feature.GetUsersFilter) ([]domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
}

func NewUsersService(usersRepository UsersRepository) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}
