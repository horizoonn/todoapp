package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM todoapp.users
	WHERE id=$1;
	`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrViolatesForeignKey) {
			return fmt.Errorf(
				"user with id='%s' has related tasks: %w",
				id,
				core_errors.ErrConflict,
			)
		}

		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id='%s': %w", id, core_errors.ErrNotFound)
	}

	return nil
}
