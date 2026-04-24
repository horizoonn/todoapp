package users_transport_http

import (
	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
)

type UserDTOResponse struct {
	ID          uuid.UUID `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000"`
	Version     int       `json:"version"        example:"3"`
	FullName    string    `json:"full_name"      example:"Ivan Ivanov"`
	PhoneNumber *string   `json:"phone_number"   example:"+79998887766"`
}

func userDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:          user.ID,
		Version:     user.Version,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
	}
}

func usersDTOFromDomains(users []domain.User) []UserDTOResponse {
	usersDTO := make([]UserDTOResponse, len(users))

	for i, user := range users {
		usersDTO[i] = userDTOFromDomain(user)
	}

	return usersDTO
}
