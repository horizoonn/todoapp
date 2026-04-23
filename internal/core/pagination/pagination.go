package pagination

import (
	"fmt"

	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

type Page struct {
	Limit  int
	Offset int
}

func Normalize(limit *int, offset *int, defaultLimit int, maxLimit int, defaultOffset int) (Page, error) {
	normalizedLimit := defaultLimit
	if limit != nil && *limit < 0 {
		return Page{}, fmt.Errorf("limit must be non-negative: %w", core_errors.ErrInvalidArgument)
	} else if limit != nil {
		normalizedLimit = min(*limit, maxLimit)
	}

	normalizedOffset := defaultOffset
	if offset != nil && *offset < 0 {
		return Page{}, fmt.Errorf("offset must be non-negative: %w", core_errors.ErrInvalidArgument)
	} else if offset != nil {
		normalizedOffset = *offset
	}

	return Page{
		Limit:  normalizedLimit,
		Offset: normalizedOffset,
	}, nil
}
