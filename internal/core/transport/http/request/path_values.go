package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

func GetIntPathValue(r *http.Request, key string) (int, error) {
	pathValue := r.PathValue(key)
	if len(pathValue) == 0 {
		return 0, fmt.Errorf("no key='%s' in path values: %w", key, core_errors.ErrInvalidArgument)
	}

	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return 0, fmt.Errorf("path value='%s' by key='%s' not a valid integer: %v: %w", pathValue, key, err, core_errors.ErrInvalidArgument)
	}

	return val, nil
}
