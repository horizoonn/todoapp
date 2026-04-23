package tasks_transport_http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_http_server "github.com/horizoonn/todoapp/internal/core/transport/http/server"
)

type TasksHTTPHandler struct {
	tasksService TasksService
}

type TasksService interface {
	CreateTask(ctx context.Context, title string, description *string, authorUserID uuid.UUID) (domain.Task, error)
	GetTasks(ctx context.Context, userID *uuid.UUID, limit *int, offset *int) ([]domain.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (domain.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	PatchTask(ctx context.Context, id uuid.UUID, patch domain.TaskPatch) (domain.Task, error)
}

func NewTasksHTTPHandler(tasksService TasksService) *TasksHTTPHandler {
	return &TasksHTTPHandler{
		tasksService: tasksService,
	}
}

func (h *TasksHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/tasks",
			Handler: h.CreateTask,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks",
			Handler: h.GetTasks,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks/{id}",
			Handler: h.GetTask,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/tasks/{id}",
			Handler: h.DeleteTask,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/tasks/{id}",
			Handler: h.PatchTask,
		},
	}
}
