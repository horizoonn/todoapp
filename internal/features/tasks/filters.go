package tasks

import "github.com/google/uuid"

type GetTasksFilter struct {
	UserID *uuid.UUID
	Limit  int
	Offset int
}

func NewGetTasksFilter(userID *uuid.UUID, limit int, offset int) GetTasksFilter {
	return GetTasksFilter{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
}
