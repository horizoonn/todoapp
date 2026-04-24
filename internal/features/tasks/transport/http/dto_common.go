package tasks_transport_http

import (
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
)

type TaskDTOResponse struct {
	ID      uuid.UUID `json:"id"                   example:"3fcef185-f1ef-4259-9bb6-e56b07129c81"`
	Version int       `json:"version"              example:"3"`

	Title       string     `json:"title"           example:"Homework"`
	Description *string    `json:"description"     example:"Finish math homework by Thursday"`
	Completed   bool       `json:"completed"       example:"false"`
	CreatedAt   time.Time  `json:"created_at"      example:"2026-02-26T10:30:00Z"`
	CompletedAt *time.Time `json:"completed_at"    example:"null"`

	AuthorUserID uuid.UUID `json:"author_user_id"  example:"ba74f37c-1ef1-458f-bd56-f7197ef440ad"`
}

func taskDTOFromDomain(task domain.Task) TaskDTOResponse {
	return TaskDTOResponse{
		ID:           task.ID,
		Version:      task.Version,
		Title:        task.Title,
		Description:  task.Description,
		Completed:    task.Completed,
		CreatedAt:    task.CreatedAt,
		CompletedAt:  task.CompletedAt,
		AuthorUserID: task.AuthorUserID,
	}
}

func tasksDTOsFromDomains(tasks []domain.Task) []TaskDTOResponse {
	dtos := make([]TaskDTOResponse, len(tasks))
	for i, task := range tasks {
		dtos[i] = taskDTOFromDomain(task)
	}

	return dtos
}
