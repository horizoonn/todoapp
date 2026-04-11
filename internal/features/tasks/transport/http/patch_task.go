package tasks_transport_http

import (
	"fmt"
	"net/http"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	core_http_types "github.com/horizoonn/todoapp/internal/core/transport/http/types"
)

type PatchTaskRequest struct {
	Title       core_http_types.Nullable[string] `json:"title"`
	Description core_http_types.Nullable[string] `json:"description"`
	Completed   core_http_types.Nullable[bool]   `json:"completed"`
}

func (r *PatchTaskRequest) Validate() error {
	if r.Title.Set {
		if r.Title.Value == nil {
			return fmt.Errorf("`Title` can't be NULL")
		}

		titleLen := len([]rune(*r.Title.Value))
		if titleLen < domain.MinTaskTitleLen || titleLen > domain.MaxTaskTitleLen {
			return fmt.Errorf("`Title` must be between %d and %d symbols", domain.MinTaskTitleLen, domain.MaxTaskTitleLen)
		}
	}

	if r.Description.Set {
		if r.Description.Value != nil {
			descriptionLen := len([]rune(*r.Description.Value))

			if descriptionLen < domain.MinTaskDescriptionLen || descriptionLen > domain.MaxTaskDescriptionLen {
				return fmt.Errorf("`Description` must be between %d and %d symbols", domain.MinTaskDescriptionLen, domain.MaxTaskDescriptionLen)
			}
		}
	}

	if r.Completed.Set {
		if r.Completed.Value == nil {
			return fmt.Errorf("`Completed` can't be NULL")
		}
	}

	return nil
}

func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get taskID path value")

		return
	}

	var request PatchTaskRequest

	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	taskPatch := taskPatchFromRequest(request)

	taskDomain, err := h.tasksService.PatchTask(ctx, taskID, taskPatch)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to patch task")

		return
	}

	response := PatchTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

type PatchTaskResponse TaskDTOResponse

func taskPatchFromRequest(request PatchTaskRequest) domain.TaskPatch {
	return domain.NewTaskPatch(
		request.Title.ToDomain(),
		request.Description.ToDomain(),
		request.Completed.ToDomain(),
	)
}
