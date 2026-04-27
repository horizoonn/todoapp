package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

func TestUserValidate(t *testing.T) {
	phone := "+79998887766"

	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
	}{
		{
			name: "valid without phone",
			user: domain.NewUser(uuid.New(), 1, "Ivan Ivanov", nil),
		},
		{
			name: "valid with phone",
			user: domain.NewUser(uuid.New(), 1, "Ivan Ivanov", &phone),
		},
		{
			name:    "full name too short",
			user:    domain.NewUser(uuid.New(), 1, "Iv", nil),
			wantErr: true,
		},
		{
			name: "phone too short",
			user: func() domain.User {
				phone := "+123"
				return domain.NewUser(uuid.New(), 1, "Ivan Ivanov", &phone)
			}(),
			wantErr: true,
		},
		{
			name: "phone must start with plus",
			user: func() domain.User {
				phone := "79998887766"
				return domain.NewUser(uuid.New(), 1, "Ivan Ivanov", &phone)
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestUserApplyPatch(t *testing.T) {
	initialPhone := "+79998887766"

	tests := []struct {
		name      string
		patch     domain.UserPatch
		assert    func(t *testing.T, user domain.User)
		wantErr   bool
		unchanged bool
	}{
		{
			name: "updates full name",
			patch: domain.NewUserPatch(
				nullableValue("Petr Petrov"),
				nullableUnset[string](),
			),
			assert: func(t *testing.T, user domain.User) {
				t.Helper()
				if user.FullName != "Petr Petrov" {
					t.Fatalf("expected full name to be patched, got %q", user.FullName)
				}
			},
		},
		{
			name: "clears phone number",
			patch: domain.NewUserPatch(
				nullableUnset[string](),
				nullableNull[string](),
			),
			assert: func(t *testing.T, user domain.User) {
				t.Helper()
				if user.PhoneNumber != nil {
					t.Fatalf("expected phone number to be cleared, got %q", *user.PhoneNumber)
				}
			},
		},
		{
			name:      "rejects empty patch",
			patch:     domain.NewUserPatch(nullableUnset[string](), nullableUnset[string]()),
			wantErr:   true,
			unchanged: true,
		},
		{
			name: "rejects null full name",
			patch: domain.NewUserPatch(
				nullableNull[string](),
				nullableUnset[string](),
			),
			wantErr:   true,
			unchanged: true,
		},
		{
			name: "rejects invalid phone and keeps original user",
			patch: domain.NewUserPatch(
				nullableUnset[string](),
				nullableValue("invalid-phone"),
			),
			wantErr:   true,
			unchanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := domain.NewUser(uuid.New(), 1, "Ivan Ivanov", &initialPhone)
			original := user

			err := user.ApplyPatch(tt.patch)
			if tt.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				if tt.unchanged && !sameUser(user, original) {
					t.Fatalf("expected user to stay unchanged, got %+v, want %+v", user, original)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tt.assert != nil {
				tt.assert(t, user)
			}
		})
	}
}

func nullableUnset[T any]() domain.Nullable[T] {
	return domain.Nullable[T]{}
}

func nullableNull[T any]() domain.Nullable[T] {
	return domain.Nullable[T]{Set: true}
}

func nullableValue[T any](value T) domain.Nullable[T] {
	return domain.Nullable[T]{
		Value: &value,
		Set:   true,
	}
}

func sameUser(a domain.User, b domain.User) bool {
	if a.ID != b.ID || a.Version != b.Version || a.FullName != b.FullName {
		return false
	}
	if a.PhoneNumber == nil || b.PhoneNumber == nil {
		return a.PhoneNumber == nil && b.PhoneNumber == nil
	}
	return *a.PhoneNumber == *b.PhoneNumber
}
