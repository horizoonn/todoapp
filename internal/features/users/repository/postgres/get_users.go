package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	users_feature "github.com/horizoonn/todoapp/internal/features/users"
)

func (r *UsersRepository) GetUsers(ctx context.Context, filter users_feature.GetUsersFilter) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, full_name, phone_number
	FROM todoapp.users
	ORDER BY full_name ASC, id ASC
	LIMIT $1
	OFFSET $2;
	`

	rows, err := r.pool.Query(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	var userModels []UserModel
	for rows.Next() {
		var userModel UserModel
		if err := userModel.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan users: %w", err)
		}

		userModels = append(userModels, userModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	userDomains := modelsToDomains(userModels)

	return userDomains, nil
}
