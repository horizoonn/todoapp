package users_transport_http

import (
	"net/http"

	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	http_utils "github.com/horizoonn/todoapp/internal/core/transport/http/utils"
)

func (h *UsersHTTPHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := http_response.NewHTTPResponseHandler(log, rw)

	userID, err := http_utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")

		return
	}

	if err := h.usersService.DeleteUser(ctx, userID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete user")

		return
	}

	responseHandler.NoContentResponse()
}
