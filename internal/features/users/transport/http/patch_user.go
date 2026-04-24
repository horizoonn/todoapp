package users_transport_http

import (
	"net/http"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	core_http_request "github.com/horizoonn/todoapp/internal/core/transport/http/request"
	core_http_response "github.com/horizoonn/todoapp/internal/core/transport/http/response"
	core_http_types "github.com/horizoonn/todoapp/internal/core/transport/http/types"
)

type PatchUserRequest struct {
	FullName    core_http_types.Nullable[string] `json:"full_name"       swaggertype:"string" example:"John Smith"`
	PhoneNumber core_http_types.Nullable[string] `json:"phone_number"    swaggertype:"string" example:"+71112223344"`
}

type PatchUserResponse UserDTOResponse

// PatchUser     godoc
// @Summary      Изменение пользователя
// @Description  Изменение информации об уже существующем в системе пользователе
// @Description  ### Логика обновления полей (Three-state logic):
// @Description  1. **Поле не передано**: `phone_number` игнорируется, значение в БД не меняется
// @Description  2. **Явно передано значение**: `"phone_number": "+711122233344"` - устанавливает новый номер телефона в БД
// @Description  3. **Передан null**: `"phone_number": null` - очищает поле в БД (set to NULL)
// @Description  Ограничения: `full_name` не может быть выставлен как null
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path string           true            "ID изменяемого пользователя" Format(uuid)
// @Param        request body PatchUserRequest true            "PatchUser тело запроса"
// @Success      200 {object} PatchUserResponse                "Успешно изменённый пользователь"
// @Failure      400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure      404 {object} core_http_response.ErrorResponse "User not found"
// @Failure      409 {object} core_http_response.ErrorResponse "Conflict"
// @Failure      500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router       /users/{id} [patch]
func (h *UsersHTTPHandler) PatchUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_http_request.GetUUIDPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")

		return
	}

	var request PatchUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	userPatch := userPatchFromRequest(request)

	userDomain, err := h.usersService.PatchUser(ctx, userID, userPatch)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to patch user")

		return
	}

	response := PatchUserResponse(userDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func userPatchFromRequest(request PatchUserRequest) domain.UserPatch {
	return domain.NewUserPatch(request.FullName.ToDomain(), request.PhoneNumber.ToDomain())
}
