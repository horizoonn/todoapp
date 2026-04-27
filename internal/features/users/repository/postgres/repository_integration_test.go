//go:build integration

package users_postgres_repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	users_feature "github.com/horizoonn/todoapp/internal/features/users"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	"github.com/horizoonn/todoapp/internal/testsupport/integration"
)

func TestUsersRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool := integration.NewPostgresPool(t)
	repository := users_postgres_repository.NewUsersRepository(pool)

	t.Run("save and get user", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		phoneNumber := "+15551234567"
		user := domain.NewUser(uuid.New(), 1, "Alice Johnson", &phoneNumber)

		savedUser, err := repository.SaveUser(ctx, user)
		if err != nil {
			t.Fatalf("save user: %v", err)
		}
		if savedUser.ID != user.ID || savedUser.FullName != user.FullName {
			t.Fatalf("expected saved user %+v, got %+v", user, savedUser)
		}

		gotUser, err := repository.GetUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		if gotUser.ID != savedUser.ID || gotUser.PhoneNumber == nil || *gotUser.PhoneNumber != phoneNumber {
			t.Fatalf("expected saved user %+v, got %+v", savedUser, gotUser)
		}
	})

	t.Run("list users sorted by full name", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		alice := domain.NewUser(uuid.New(), 1, "Alice Johnson", nil)
		aaron := domain.NewUser(uuid.New(), 1, "Aaron Smith", nil)

		if _, err := repository.SaveUser(ctx, alice); err != nil {
			t.Fatalf("save alice: %v", err)
		}
		if _, err := repository.SaveUser(ctx, aaron); err != nil {
			t.Fatalf("save aaron: %v", err)
		}

		users, err := repository.GetUsers(ctx, users_feature.NewGetUsersFilter(10, 0))
		if err != nil {
			t.Fatalf("get users: %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("expected 2 users, got %+v", users)
		}
		if users[0].ID != aaron.ID || users[1].ID != alice.ID {
			t.Fatalf("expected users sorted by full_name, got %+v", users)
		}
	})

	t.Run("update user increments version", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		user := domain.NewUser(uuid.New(), 1, "Alice Johnson", nil)
		savedUser, err := repository.SaveUser(ctx, user)
		if err != nil {
			t.Fatalf("save user: %v", err)
		}

		updatedPhoneNumber := "+15557654321"
		savedUser.FullName = "Alice Cooper"
		savedUser.PhoneNumber = &updatedPhoneNumber
		updatedUser, err := repository.UpdateUser(ctx, savedUser)
		if err != nil {
			t.Fatalf("update user: %v", err)
		}
		if updatedUser.Version != savedUser.Version+1 {
			t.Fatalf("expected version %d, got %d", savedUser.Version+1, updatedUser.Version)
		}
		if updatedUser.FullName != savedUser.FullName || updatedUser.PhoneNumber == nil || *updatedUser.PhoneNumber != updatedPhoneNumber {
			t.Fatalf("expected updated user %+v, got %+v", savedUser, updatedUser)
		}
	})

	t.Run("stale update returns conflict", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		user := domain.NewUser(uuid.New(), 1, "Alice Johnson", nil)
		savedUser, err := repository.SaveUser(ctx, user)
		if err != nil {
			t.Fatalf("save user: %v", err)
		}
		if _, err := repository.UpdateUser(ctx, savedUser); err != nil {
			t.Fatalf("update user first time: %v", err)
		}

		_, err = repository.UpdateUser(ctx, savedUser)
		if !errors.Is(err, core_errors.ErrConflict) {
			t.Fatalf("expected ErrConflict for stale version update, got %v", err)
		}
	})

	t.Run("delete user and report not found", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		user := domain.NewUser(uuid.New(), 1, "Alice Johnson", nil)
		savedUser, err := repository.SaveUser(ctx, user)
		if err != nil {
			t.Fatalf("save user: %v", err)
		}

		if err := repository.DeleteUser(ctx, savedUser.ID); err != nil {
			t.Fatalf("delete user: %v", err)
		}

		_, err = repository.GetUser(ctx, savedUser.ID)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound after delete, got %v", err)
		}

		err = repository.DeleteUser(ctx, savedUser.ID)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for deleting missing user, got %v", err)
		}
	})
}
