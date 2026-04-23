package tasks_postgres_repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/horizoonn/todoapp/internal/core/domain"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
)

func (r *TasksRepository) GetTasks(ctx context.Context, filter tasks_feature.GetTasksFilter) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`
		SELECT id, version, title, description, completed, created_at, completed_at, author_user_id
		FROM todoapp.tasks
	`)

	args := []any{filter.Limit, filter.Offset}

	if filter.UserID != nil {
		queryBuilder.WriteString(" WHERE author_user_id=$3")
		args = append(args, *filter.UserID)
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC, id ASC LIMIT $1 OFFSET $2;")

	query := queryBuilder.String()

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var taskModels []TaskModel
	for rows.Next() {
		var taskModel TaskModel
		if err := taskModel.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan tasks: %w", err)
		}

		taskModels = append(taskModels, taskModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	taskDomains := modelsToDomains(taskModels)

	return taskDomains, nil
}
