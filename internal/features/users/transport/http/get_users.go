package users_transport_http

import (
	"fmt"
	"net/http"

	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	http_utils "github.com/horizoonn/todoapp/internal/core/transport/http/utils"
)

type GetUsersResponse []UserDTOResponse

func (h *UsersHTTPHandler) GetUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := http_response.NewHTTPResponseHandler(log, rw)

	limit, offset, err := getLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get 'limit'/'offset' query param")
		return
	}

	userDomains, err := h.usersService.GetUsers(ctx, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get users")
		return
	}

	response := GetUsersResponse(usersDTOFromDomains(userDomains))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func getLimitOffsetQueryParams(r *http.Request) (*int, *int, error) {
	limit, err := http_utils.GetIntQueryParam(r, "limit")
	if err != nil {
		return nil, nil, fmt.Errorf("get 'limit' query param: %w", err)
	}

	offset, err := http_utils.GetIntQueryParam(r, "offset")
	if err != nil {
		return nil, nil, fmt.Errorf("get 'offset' query param: %w", err)
	}

	return limit, offset, nil
}
