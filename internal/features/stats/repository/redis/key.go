package stats_redis_repository

import (
	"fmt"

	"github.com/google/uuid"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

func statsKey(filter stats_feature.GetStatsFilter, version int64) string {
	userID := "all"
	if filter.UserID != nil {
		userID = filter.UserID.String()
	}

	from := "all"
	if filter.From != nil {
		from = filter.From.UTC().Format(dateTimeLayout)
	}

	to := "all"
	if filter.To != nil {
		to = filter.To.UTC().Format(dateTimeLayout)
	}

	return fmt.Sprintf("stats:version:%d:user_id:%s:from:%s:to:%s", version, userID, from, to)
}

func statsVersionAllKey() string {
	return "stats:version:all"
}

func statsVersionUserKey(userID uuid.UUID) string {
	return fmt.Sprintf("stats:version:user_id:%s", userID.String())
}
