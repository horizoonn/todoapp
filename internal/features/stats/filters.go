package stats

import (
	"time"

	"github.com/google/uuid"
)

type GetStatsFilter struct {
	UserID *uuid.UUID
	From   *time.Time
	To     *time.Time
}

func NewGetStatsFilter(userID *uuid.UUID, from *time.Time, to *time.Time) GetStatsFilter {
	return GetStatsFilter{
		UserID: userID,
		From:   from,
		To:     to,
	}
}
