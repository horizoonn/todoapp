package tasks_transport_http

import (
	"net/http"

	"github.com/google/uuid"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	Title        string    `json:"title"`
	Description  *string   `json:"description"`
	AuthorUserID uuid.UUID `json:"author_user_id"`
}

type CreateTaskResponse TaskDTOResponse

func (h *TasksHTTPHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request CreateTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	taskDomain, err := h.tasksService.CreateTask(ctx, request.Title, request.Description, request.AuthorUserID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create task")

		return
	}

	response := CreateTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}
