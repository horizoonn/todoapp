package tasks_transport_http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	tasks_transport_http "github.com/horizoonn/todoapp/internal/features/tasks/transport/http"
	tasks_transport_http_mocks "github.com/horizoonn/todoapp/internal/features/tasks/transport/http/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreateTaskReturnsCreatedTask(t *testing.T) {
	task := newTask()
	ctrl := gomock.NewController(t)

	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	service.EXPECT().
		CreateTask(gomock.Any(), "Homework", gomock.Any(), task.AuthorUserID).
		DoAndReturn(func(_ context.Context, title string, description *string, authorUserID uuid.UUID) (domain.Task, error) {
			if title != "Homework" {
				t.Fatalf("expected title to be passed to service, got %q", title)
			}
			if description == nil || *description != "Finish math homework" {
				t.Fatalf("expected description to be passed to service, got %v", description)
			}
			if authorUserID != task.AuthorUserID {
				t.Fatalf("expected author user id %s, got %s", task.AuthorUserID, authorUserID)
			}
			return task, nil
		})

	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	requestBody := `{"title":"Homework","description":"Finish math homework","author_user_id":"` + task.AuthorUserID.String() + `"}`
	request := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	handler.CreateTask(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response tasks_transport_http.CreateTaskResponse
	decodeJSONResponse(t, recorder, &response)
	if response.ID != task.ID || response.AuthorUserID != task.AuthorUserID {
		t.Fatalf("expected created task %+v, got %+v", task, response)
	}
}

func TestCreateTaskReturnsBadRequestForInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":`))
	recorder := httptest.NewRecorder()

	handler.CreateTask(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetTasksParsesQueryAndReturnsTasks(t *testing.T) {
	task := newTask()
	limit := 10
	offset := 5
	ctrl := gomock.NewController(t)

	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	service.EXPECT().
		GetTasks(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, userID *uuid.UUID, gotLimit *int, gotOffset *int) ([]domain.Task, error) {
			if userID == nil || *userID != task.AuthorUserID {
				t.Fatalf("expected user id %s, got %v", task.AuthorUserID, userID)
			}
			if gotLimit == nil || *gotLimit != limit {
				t.Fatalf("expected limit %d, got %v", limit, gotLimit)
			}
			if gotOffset == nil || *gotOffset != offset {
				t.Fatalf("expected offset %d, got %v", offset, gotOffset)
			}
			return []domain.Task{task}, nil
		})

	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/tasks?user_id="+task.AuthorUserID.String()+"&limit=10&offset=5", nil)
	recorder := httptest.NewRecorder()

	handler.GetTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	var response tasks_transport_http.GetTasksResponse
	decodeJSONResponse(t, recorder, &response)
	if len(response) != 1 || response[0].ID != task.ID {
		t.Fatalf("expected task id %s, got %+v", task.ID, response)
	}
}

func TestGetTasksReturnsBadRequestForInvalidUserIDQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/tasks?user_id=bad", nil)
	recorder := httptest.NewRecorder()

	handler.GetTasks(recorder, request)

	assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_argument")
}

func TestGetTaskReturnsNotFoundForMissingTask(t *testing.T) {
	taskID := uuid.New()
	ctrl := gomock.NewController(t)

	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	service.EXPECT().
		GetTask(gomock.Any(), taskID).
		Return(domain.Task{}, core_errors.ErrNotFound)

	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String(), nil)
	request.SetPathValue("id", taskID.String())
	recorder := httptest.NewRecorder()

	handler.GetTask(recorder, request)

	assertErrorResponse(t, recorder, http.StatusNotFound, "not_found")
}

func TestPatchTaskParsesNullableFields(t *testing.T) {
	task := newTask()
	completed := true
	ctrl := gomock.NewController(t)

	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	service.EXPECT().
		PatchTask(gomock.Any(), task.ID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, patch domain.TaskPatch) (domain.Task, error) {
			if patch.Title.Set {
				t.Fatalf("expected title patch to be unset, got %+v", patch.Title)
			}
			if !patch.Description.Set || patch.Description.Value != nil {
				t.Fatalf("expected description patch to be explicit null, got %+v", patch.Description)
			}
			if !patch.Completed.Set || patch.Completed.Value == nil || *patch.Completed.Value != completed {
				t.Fatalf("expected completed patch to be %v, got %+v", completed, patch.Completed)
			}
			return task, nil
		})

	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodPatch, "/tasks/"+task.ID.String(), bytes.NewBufferString(`{"description":null,"completed":true}`))
	request.SetPathValue("id", task.ID.String())
	recorder := httptest.NewRecorder()

	handler.PatchTask(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

func TestDeleteTaskReturnsNoContent(t *testing.T) {
	taskID := uuid.New()
	ctrl := gomock.NewController(t)

	service := tasks_transport_http_mocks.NewMockTasksService(ctrl)
	service.EXPECT().
		DeleteTask(gomock.Any(), taskID).
		Return(nil)

	handler := tasks_transport_http.NewTasksHTTPHandler(service)
	request := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	request.SetPathValue("id", taskID.String())
	recorder := httptest.NewRecorder()

	handler.DeleteTask(recorder, request)

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

func newTask() domain.Task {
	description := "Finish math homework"
	return domain.NewTask(
		uuid.New(),
		1,
		"Homework",
		&description,
		false,
		time.Date(2024, 4, 25, 10, 0, 0, 0, time.UTC),
		nil,
		uuid.New(),
	)
}
