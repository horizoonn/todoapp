package users_service

import (
	"context"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	"github.com/horizoonn/todoapp/internal/core/pagination"
	users_feature "github.com/horizoonn/todoapp/internal/features/users"
)

const (
	defaultUsersLimit  = 100
	maxUsersLimit      = 100
	defaultUsersOffset = 0
)

func (s *UsersService) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	page, err := pagination.Normalize(limit, offset, defaultUsersLimit, maxUsersLimit, defaultUsersOffset)
	if err != nil {
		return nil, fmt.Errorf("normalize pagination: %w", err)
	}

	filter := users_feature.NewGetUsersFilter(page.Limit, page.Offset)

	users, err := s.usersRepository.GetUsers(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get users from repository: %w", err)
	}

	return users, nil
}
