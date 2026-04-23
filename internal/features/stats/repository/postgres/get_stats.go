package stats_postgres_repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
)

func (r *StatsRepository) GetStats(ctx context.Context, filter stats_feature.GetStatsFilter) (domain.Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`
		WITH stats AS (
			SELECT
				COUNT(*) AS tasks_created,
				COUNT(*) FILTER (WHERE completed = TRUE) AS tasks_completed,
				AVG(completed_at - created_at) FILTER (WHERE completed = TRUE) AS tasks_average_completion_time
			FROM todoapp.tasks
		`)

	args := []any{}
	conditions := []string{}

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("author_user_id=$%d", len(args)+1))
		args = append(args, *filter.UserID)
	}

	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at>=$%d", len(args)+1))
		args = append(args, *filter.From)
	}

	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at<$%d", len(args)+1))
		args = append(args, *filter.To)
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE " + strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(`
		)
		SELECT
			tasks_created,
			tasks_completed,
			CASE
				WHEN tasks_created = 0 THEN NULL
				ELSE tasks_completed::DOUBLE PRECISION / tasks_created::DOUBLE PRECISION * 100
			END AS tasks_completed_rate,
			EXTRACT(EPOCH FROM tasks_average_completion_time)::DOUBLE PRECISION AS tasks_average_completion_time_seconds
		FROM stats;
	`)

	row := r.pool.QueryRow(ctx, queryBuilder.String(), args...)

	var statsModel StatsModel
	if err := statsModel.Scan(row); err != nil {
		return domain.Stats{}, fmt.Errorf("scan stats: %w", err)
	}

	return modelToDomain(statsModel), nil
}
