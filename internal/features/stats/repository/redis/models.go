package stats_redis_repository

import (
	"time"

	"github.com/horizoonn/todoapp/internal/core/domain"
)

const dateTimeLayout = time.RFC3339Nano

type StatsModel struct {
	TasksCreated                    int      `json:"tasks_created"`
	TasksCompleted                  int      `json:"tasks_completed"`
	TasksCompletedRate              *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTimeNanos *int64   `json:"tasks_average_completion_time_nanos"`
}

func modelToDomain(model StatsModel) domain.Stats {
	var averageCompletionTime *time.Duration
	if model.TasksAverageCompletionTimeNanos != nil {
		duration := time.Duration(*model.TasksAverageCompletionTimeNanos)
		averageCompletionTime = &duration
	}

	return domain.NewStats(
		model.TasksCreated,
		model.TasksCompleted,
		model.TasksCompletedRate,
		averageCompletionTime,
	)
}

func modelFromDomain(stats domain.Stats) StatsModel {
	var averageCompletionTimeNanos *int64
	if stats.TasksAverageCompletionTime != nil {
		nanos := int64(*stats.TasksAverageCompletionTime)
		averageCompletionTimeNanos = &nanos
	}

	return StatsModel{
		TasksCreated:                    stats.TasksCreated,
		TasksCompleted:                  stats.TasksCompleted,
		TasksCompletedRate:              stats.TasksCompletedRate,
		TasksAverageCompletionTimeNanos: averageCompletionTimeNanos,
	}
}
