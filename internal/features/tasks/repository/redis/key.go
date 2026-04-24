package tasks_redis_repository

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

func taskKey(id uuid.UUID) string {
	return fmt.Sprintf("tasks:item:id:%s", id.String())
}

func tasksListKey(filter tasks_feature.GetTasksFilter, version int64) string {
	userID := "all"
	if filter.UserID != nil {
		userID = filter.UserID.String()
	}

	return fmt.Sprintf(
		"tasks:list:version:%s:user_id:%s:limit:%d:offset:%d",
		strconv.FormatInt(version, 10),
		userID,
		filter.Limit,
		filter.Offset,
	)
}

func tasksListVersionAllKey() string {
	return "tasks:list:version:all"
}

func tasksListVersionUserKey(userID uuid.UUID) string {
	return fmt.Sprintf("tasks:list:version:user_id:%s", userID.String())
}
