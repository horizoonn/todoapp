package users_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	users_feature "github.com/horizoonn/todoapp/internal/features/users"
	users_service "github.com/horizoonn/todoapp/internal/features/users/service"
	users_service_mocks "github.com/horizoonn/todoapp/internal/features/users/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreateUserSavesValidUser(t *testing.T) {
	ctx := context.Background()
	phoneNumber := "+15551234567"
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		SaveUser(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, user domain.User) (domain.User, error) {
			user.Version++
			return user, nil
		})

	service := users_service.NewUsersService(repository)

	user, err := service.CreateUser(ctx, "Alice Johnson", &phoneNumber)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user.FullName != "Alice Johnson" {
		t.Fatalf("expected full name %q, got %q", "Alice Johnson", user.FullName)
	}
	if user.PhoneNumber == nil || *user.PhoneNumber != phoneNumber {
		t.Fatalf("expected phone number %q, got %v", phoneNumber, user.PhoneNumber)
	}
	if user.Version != 2 {
		t.Fatalf("expected repository result version 2, got %d", user.Version)
	}
}

func TestCreateUserValidationFailureDoesNotSave(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	service := users_service.NewUsersService(repository)

	_, err := service.CreateUser(ctx, "Al", nil)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetUsersNormalizesPaginationAndLoadsRepository(t *testing.T) {
	ctx := context.Background()
	limit := 10
	offset := 5
	users := []domain.User{newUser()}
	wantFilter := users_feature.NewGetUsersFilter(limit, offset)
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUsers(ctx, wantFilter).
		Return(users, nil)

	service := users_service.NewUsersService(repository)

	got, err := service.GetUsers(ctx, &limit, &offset)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != users[0].ID {
		t.Fatalf("expected repository users %+v, got %+v", users, got)
	}
}

func TestGetUsersUsesDefaultPagination(t *testing.T) {
	ctx := context.Background()
	wantFilter := users_feature.NewGetUsersFilter(100, 0)
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUsers(ctx, wantFilter).
		Return(nil, nil)

	service := users_service.NewUsersService(repository)

	_, err := service.GetUsers(ctx, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestGetUsersRejectsInvalidPaginationBeforeRepository(t *testing.T) {
	ctx := context.Background()
	limit := -1
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	service := users_service.NewUsersService(repository)

	_, err := service.GetUsers(ctx, &limit, nil)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetUserLoadsRepository(t *testing.T) {
	ctx := context.Background()
	user := newUser()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUser(ctx, user.ID).
		Return(user, nil)

	service := users_service.NewUsersService(repository)

	got, err := service.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, got.ID)
	}
}

func TestGetUserPropagatesRepositoryError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUser(ctx, userID).
		Return(domain.User{}, core_errors.ErrNotFound)

	service := users_service.NewUsersService(repository)

	_, err := service.GetUser(ctx, userID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPatchUserUpdatesPatchedUser(t *testing.T) {
	ctx := context.Background()
	user := newUser()
	updatedFullName := "Bob Smith"
	updatedUser := user
	updatedUser.FullName = updatedFullName
	patchedUser := updatedUser
	patchedUser.Version++
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	gomock.InOrder(
		repository.EXPECT().
			GetUser(ctx, user.ID).
			Return(user, nil),
		repository.EXPECT().
			UpdateUser(ctx, updatedUser).
			Return(patchedUser, nil),
	)

	service := users_service.NewUsersService(repository)

	got, err := service.PatchUser(ctx, user.ID, domain.NewUserPatch(
		domain.Nullable[string]{Value: &updatedFullName, Set: true},
		domain.Nullable[string]{},
	))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.FullName != updatedFullName {
		t.Fatalf("expected patched full name %q, got %q", updatedFullName, got.FullName)
	}
	if got.Version != user.Version+1 {
		t.Fatalf("expected repository result version %d, got %d", user.Version+1, got.Version)
	}
}

func TestPatchUserValidationFailureDoesNotUpdate(t *testing.T) {
	ctx := context.Background()
	user := newUser()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUser(ctx, user.ID).
		Return(user, nil)

	service := users_service.NewUsersService(repository)

	_, err := service.PatchUser(ctx, user.ID, domain.NewUserPatch(
		domain.Nullable[string]{},
		domain.Nullable[string]{},
	))
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestPatchUserPropagatesGetErrorBeforeUpdate(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	updatedFullName := "Bob Smith"
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		GetUser(ctx, userID).
		Return(domain.User{}, core_errors.ErrNotFound)

	service := users_service.NewUsersService(repository)

	_, err := service.PatchUser(ctx, userID, domain.NewUserPatch(
		domain.Nullable[string]{Value: &updatedFullName, Set: true},
		domain.Nullable[string]{},
	))
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteUserDeletesRepositoryUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		DeleteUser(ctx, userID).
		Return(nil)

	service := users_service.NewUsersService(repository)

	if err := service.DeleteUser(ctx, userID); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDeleteUserPropagatesRepositoryError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ctrl := gomock.NewController(t)

	repository := users_service_mocks.NewMockUsersRepository(ctrl)
	repository.EXPECT().
		DeleteUser(ctx, userID).
		Return(core_errors.ErrNotFound)

	service := users_service.NewUsersService(repository)

	err := service.DeleteUser(ctx, userID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func newUser() domain.User {
	phoneNumber := "+15551234567"
	return domain.NewUser(uuid.New(), 1, "Alice Johnson", &phoneNumber)
}
