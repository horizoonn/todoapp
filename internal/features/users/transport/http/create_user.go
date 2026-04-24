package users_transport_http

import (
	"net/http"

	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
)

type CreateUserRequest struct {
	FullName    string  `json:"full_name"         example:"Ivan Ivanov"`
	PhoneNumber *string `json:"phone_number"      example:"+79998887766"`
}

type CreateUserResponse UserDTOResponse

// CreateUser   godoc
// @Summary     Создать пользователя
// @Description Создать нового пользователя в системе
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body     CreateUserRequest  true "CreateUser тело запроса"
// @Success     201     {object} CreateUserResponse "Успешно созданный пользователь"
// @Failure     400     {object} core_http_response.ErrorResponse "Bad request"
// @Failure     500     {object} core_http_response.ErrorResponse "Internal server error"
// @Router      /users [post]
func (h *UsersHTTPHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request CreateUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	userDomain, err := h.usersService.CreateUser(ctx, request.FullName, request.PhoneNumber)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create user")

		return
	}

	response := CreateUserResponse(userDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}
