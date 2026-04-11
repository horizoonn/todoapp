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

	stats := domain.CalcStats(tasks)

	return stats, nil
}
