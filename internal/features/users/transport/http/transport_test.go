package users_transport_http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	users_transport_http "github.com/horizoonn/todoapp/internal/features/users/transport/http"
	users_transport_http_mocks "github.com/horizoonn/todoapp/internal/features/users/transport/http/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreateUserReturnsCreatedUser(t *testing.T) {
	phoneNumber := "+15551234567"
	user := domain.NewUser(uuid.New(), 1, "Alice Johnson", &phoneNumber)
	ctrl := gomock.NewController(t)

	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	service.EXPECT().
		CreateUser(gomock.Any(), "Alice Johnson", gomock.Any()).
		DoAndReturn(func(_ context.Context, fullName string, gotPhoneNumber *string) (domain.User, error) {
			if fullName != "Alice Johnson" {
				t.Fatalf("expected full name to be passed to service, got %q", fullName)
			}
			if gotPhoneNumber == nil || *gotPhoneNumber != phoneNumber {
				t.Fatalf("expected phone number %q to be passed to service, got %v", phoneNumber, gotPhoneNumber)
			}
			return user, nil
		})

	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"full_name":"Alice Johnson","phone_number":"+15551234567"}`))
	recorder := httptest.NewRecorder()

	handler.CreateUser(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response users_transport_http.CreateUserResponse
	decodeJSONResponse(t, recorder, &response)
	if response.ID != user.ID || response.FullName != user.FullName {
		t.Fatalf("expected created user %+v, got %+v", user, response)
	}
}

func TestCreateUserReturnsBadRequestForInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"full_name":`))
	recorder := httptest.NewRecorder()

	handler.CreateUser(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetUsersParsesQueryAndReturnsUsers(t *testing.T) {
	limit := 10
	offset := 5
	users := []domain.User{newUser()}
	ctrl := gomock.NewController(t)

	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	service.EXPECT().
		GetUsers(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, gotLimit *int, gotOffset *int) ([]domain.User, error) {
			if gotLimit == nil || *gotLimit != limit {
				t.Fatalf("expected limit %d, got %v", limit, gotLimit)
			}
			if gotOffset == nil || *gotOffset != offset {
				t.Fatalf("expected offset %d, got %v", offset, gotOffset)
			}
			return users, nil
		})

	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/users?limit=10&offset=5", nil)
	recorder := httptest.NewRecorder()

	handler.GetUsers(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response users_transport_http.GetUsersResponse
	decodeJSONResponse(t, recorder, &response)
	if len(response) != 1 || response[0].ID != users[0].ID {
		t.Fatalf("expected users %+v, got %+v", users, response)
	}
}

func TestGetUsersReturnsBadRequestForInvalidQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/users?limit=bad", nil)
	recorder := httptest.NewRecorder()

	handler.GetUsers(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetUserReturnsBadRequestForInvalidPathID(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	request.SetPathValue("id", "not-a-uuid")
	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetUserReturnsNotFoundForMissingUser(t *testing.T) {
	userID := uuid.New()
	ctrl := gomock.NewController(t)

	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	service.EXPECT().
		GetUser(gomock.Any(), userID).
		Return(domain.User{}, core_errors.ErrNotFound)

	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	request.SetPathValue("id", userID.String())
	recorder := httptest.NewRecorder()

	handler.GetUser(recorder, request)

	assertErrorResponse(t, recorder, http.StatusNotFound, "not_found")
}

func TestPatchUserParsesNullableFields(t *testing.T) {
	userID := uuid.New()
	user := domain.NewUser(userID, 2, "Alice Johnson", nil)
	ctrl := gomock.NewController(t)

	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	service.EXPECT().
		PatchUser(gomock.Any(), userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, patch domain.UserPatch) (domain.User, error) {
			if patch.FullName.Set {
				t.Fatalf("expected full_name patch to be unset, got %+v", patch.FullName)
			}
			if !patch.PhoneNumber.Set || patch.PhoneNumber.Value != nil {
				t.Fatalf("expected phone_number patch to be explicit null, got %+v", patch.PhoneNumber)
			}
			return user, nil
		})

	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodPatch, "/users/"+userID.String(), bytes.NewBufferString(`{"phone_number":null}`))
	request.SetPathValue("id", userID.String())
	recorder := httptest.NewRecorder()

	handler.PatchUser(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestDeleteUserReturnsNoContent(t *testing.T) {
	userID := uuid.New()
	ctrl := gomock.NewController(t)

	service := users_transport_http_mocks.NewMockUsersService(ctrl)
	service.EXPECT().
		DeleteUser(gomock.Any(), userID).
		Return(nil)

	handler := users_transport_http.NewUsersHTTPHandler(service)
	request := httptest.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	request.SetPathValue("id", userID.String())
	recorder := httptest.NewRecorder()

	handler.DeleteUser(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, statusCode int, code string) {
	t.Helper()

	if recorder.Code != statusCode {
		t.Fatalf("expected status %d, got %d with body %s", statusCode, recorder.Code, recorder.Body.String())
	}

	var response core_http_response.ErrorResponse
	decodeJSONResponse(t, recorder, &response)
	if response.Code != code {
		t.Fatalf("expected error code %q, got %q", code, response.Code)
	}
}

func decodeJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, dest any) {
	t.Helper()

	if err := json.NewDecoder(recorder.Body).Decode(dest); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

func newUser() domain.User {
	phoneNumber := "+15551234567"
	return domain.NewUser(uuid.New(), 1, "Alice Johnson", &phoneNumber)
}
