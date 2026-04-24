package tasks_transport_http

import (
	"net/http"

	"github.com/google/uuid"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	Title        string    `json:"title"           example:"Homework"`
	Description  *string   `json:"description"     example:"Finish math homework by Thursday"  `
	AuthorUserID uuid.UUID `json:"author_user_id"  example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateTaskResponse TaskDTOResponse

// CreateTask    godoc
// @Summary      Создать задачу
// @Description  Создать новую задачу в системе
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        request  body      CreateTaskRequest   true  "CreateTask тело запроса"
// @Success      201      {object}  CreateTaskResponse  "Успешно созданная задача"
// @Failure      400      {object}  core_http_response.ErrorResponse "Bad request or author not found"
// @Failure      500      {object}  core_http_response.ErrorResponse "Internal server error"
// @Router       /tasks [post]
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
