package core_http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_logger "github.com/horizoonn/todoapp/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *core_logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(responseBody any, statusCode int) {
	h.rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) HTMLResponse(htmlFile domain.File) {
	h.rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.rw.WriteHeader(http.StatusOK)

	if _, err := h.rw.Write(htmlFile.Buffer()); err != nil {
		h.log.Error("write HTML HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode    int
		publicMessage string
		code          string
		logFunc       func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		publicMessage = "invalid request"
		code = "invalid_argument"
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		publicMessage = "resource not found"
		code = "not_found"
		logFunc = h.log.Debug

	case errors.Is(err, core_errors.ErrConflict):
		statusCode = http.StatusConflict
		publicMessage = "resource conflict"
		code = "conflict"
		logFunc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		publicMessage = "internal server error"
		code = "internal_error"
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	h.errorResponse(statusCode, publicMessage, code)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	h.errorResponse(statusCode, "internal server error", "internal_error")
}

func (h *HTTPResponseHandler) errorResponse(
	statusCode int,
	message string,
	code string,
) {
	response := ErrorResponse{
		Message: message,
		Code:    code,
	}

	h.JSONResponse(response, statusCode)
}
