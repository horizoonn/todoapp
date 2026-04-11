package stats_service

import (
	"context"
	"fmt"
	"time"

	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

func (s *StatsService) GetStats(ctx context.Context, userID *int, from *time.Time, to *time.Time) (domain.Stats, error) {
	if from != nil && to != nil {
		if to.Before(*from) || to.Equal(*from) {
			return domain.Stats{}, fmt.Errorf("`to` must be after `from`: %w", core_errors.ErrInvalidArgument)
		}
	}

	tasks, err := s.statsRepository.GetTasks(ctx, userID, from, to)
	if err != nil {
		return domain.Stats{}, fmt.Errorf("get tasks from repository: %w", err)
	}

	stats := calcStats(tasks)

	return stats, nil
}

func calcStats(tasks []domain.Task) domain.Stats {
	if len(tasks) == 0 {
		return domain.NewStats(0, 0, nil, nil)
	}

	tasksCreated := len(tasks)

	tasksCompleted := 0
	var totalCompletionDuration time.Duration

	for _, task := range tasks {
		if task.Completed {
			tasksCompleted++
		}

		completionDuration := task.CompletionDuration()
		if completionDuration != nil {
			totalCompletionDuration += *completionDuration
		}
	}

	tasksCompletedRate := float64(tasksCompleted) / float64(tasksCreated) * 100

	var tasksAverageCompletionTime *time.Duration
	if tasksCompleted > 0 {
		avg := totalCompletionDuration / time.Duration(tasksCompleted)

		tasksAverageCompletionTime = &avg
	}

	return domain.NewStats(tasksCompleted, tasksCompleted, &tasksCompletedRate, tasksAverageCompletionTime)
}
