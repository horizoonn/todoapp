package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	core_postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE todoapp.tasks
	SET 
		title=$1,
		description=$2,
		completed=$3,
		completed_at=$4,
		version=version + 1
	WHERE id=$5 AND version=$6
	RETURNING 
		id, version, title, description, completed, created_at, completed_at, author_user_id;
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Completed, task.CompletedAt, task.ID, task.Version)

	var taskModel TaskModel
	if err := taskModel.Scan(row); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf(
				"task with id='%s' concurrently accessed: %w",
				task.ID,
				core_errors.ErrConflict,
			)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := modelToDomain(taskModel)

	return taskDomain, nil
}
