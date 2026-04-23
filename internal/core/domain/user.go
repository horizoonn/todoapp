package domain

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

const (
	MinUserFullNameLen = 3
	MaxUserFullNameLen = 100

	MinUserPhoneNumberLen = 10
	MaxUserPhoneNumberLen = 15
)

var phoneRegexp = regexp.MustCompile(`^\+[0-9]{9,14}$`)

type User struct {
	ID      uuid.UUID
	Version int

	FullName    string
	PhoneNumber *string
}

func NewUser(id uuid.UUID, version int, fullName string, phoneNumber *string) User {
	return User{
		ID:          id,
		Version:     version,
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}
}

func CreateUser(
	fullName string,
	phoneNumber *string,
) User {
	var (
		id      = uuid.New()
		version = 1
	)

	return NewUser(
		id,
		version,
		fullName,
		phoneNumber,
	)
}

func (u *User) Validate() error {
	fullNameLen := len([]rune(u.FullName))
	if fullNameLen < MinUserFullNameLen || fullNameLen > MaxUserFullNameLen {
		return fmt.Errorf("invalid `FullName` len: %d (min: %d, max: %d): %w", fullNameLen, MinUserFullNameLen, MaxUserFullNameLen, core_errors.ErrInvalidArgument)
	}

	if u.PhoneNumber != nil {
		phoneNumberLen := len([]rune(*u.PhoneNumber))
		if phoneNumberLen < MinUserPhoneNumberLen || phoneNumberLen > MaxUserPhoneNumberLen {
			return fmt.Errorf("invalid `PhoneNumber` len: %d (min: %d, max: %d): %w", phoneNumberLen, MinUserPhoneNumberLen, MaxUserPhoneNumberLen, core_errors.ErrInvalidArgument)
		}

		if !phoneRegexp.MatchString(*u.PhoneNumber) {
			return fmt.Errorf("invalid `Phone Number` format: must start with '+' followed by 9-14 digits: %w", core_errors.ErrInvalidArgument)
		}
	}

	return nil
}

type UserPatch struct {
	FullName    Nullable[string]
	PhoneNumber Nullable[string]
}

func NewUserPatch(fullName Nullable[string], phoneNumber Nullable[string]) UserPatch {
	return UserPatch{
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}
}

func (p *UserPatch) Validate() error {
	if !p.FullName.Set && !p.PhoneNumber.Set {
		return fmt.Errorf("user patch must contain at least one field: %w", core_errors.ErrInvalidArgument)
	}

	if p.FullName.Set && p.FullName.Value == nil {
		return fmt.Errorf("`FullName` can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (u *User) ApplyPatch(patch UserPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate user patch: %w", err)
	}

	tmp := *u

	if patch.FullName.Set {
		tmp.FullName = *patch.FullName.Value
	}

	if patch.PhoneNumber.Set {
		tmp.PhoneNumber = patch.PhoneNumber.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched user: %w", err)
	}

	*u = tmp

	return nil
}
