package stats_postgres_repository

import (
	"time"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_postgres_pool "github.com/horizoonn/todoapp/internal/core/repository/postgres/pool"
)

type StatsModel struct {
	TasksCreated                       int64
	TasksCompleted                     int64
	TasksCompletedRate                 *float64
	TasksAverageCompletionTimeSeconds  *float64
	TasksAverageCompletionTimeDuration *time.Duration
}

func (m *StatsModel) Scan(row core_postgres_pool.Row) error {
	if err := row.Scan(
		&m.TasksCreated,
		&m.TasksCompleted,
		&m.TasksCompletedRate,
		&m.TasksAverageCompletionTimeSeconds,
	); err != nil {
		return err
	}

	if m.TasksAverageCompletionTimeSeconds != nil {
		duration := time.Duration(*m.TasksAverageCompletionTimeSeconds * float64(time.Second))
		m.TasksAverageCompletionTimeDuration = &duration
	}

	return nil
}

func modelToDomain(model StatsModel) domain.Stats {
	return domain.NewStats(
		int(model.TasksCreated),
		int(model.TasksCompleted),
		model.TasksCompletedRate,
		model.TasksAverageCompletionTimeDuration,
	)
}
