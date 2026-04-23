package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

func GetUUIDQueryParam(r *http.Request, key string) (*uuid.UUID, error) {
	param := r.URL.Query().Get(key)
	if len(param) == 0 {
		return nil, nil
	}

	val, err := uuid.Parse(param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid uuid: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &val, nil
}

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	param := r.URL.Query().Get(key)
	if len(param) == 0 {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid integer: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &val, nil
}

func GetDateQueryParam(r *http.Request, key string) (*time.Time, error) {
	param := r.URL.Query().Get(key)
	if len(param) == 0 {
		return nil, nil
	}

	layout := "2006-01-02"

	date, err := time.Parse(layout, param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid date: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &date, nil
}
