//go:build integration

package stats_redis_repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
	stats_redis_repository "github.com/horizoonn/todoapp/internal/features/stats/repository/redis"
	"github.com/horizoonn/todoapp/internal/testsupport/integration"
)

func TestStatsRedisRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool := integration.NewRedisPool(t)
	repository := stats_redis_repository.NewStatsRepository(pool)

	t.Run("set and get stats", func(t *testing.T) {
		integration.FlushRedis(t, pool)

		filter := newStatsFilter(uuid.New())
		completedRate := 75.0
		averageCompletionTime := 2 * time.Hour
		stats := domain.NewStats(4, 3, &completedRate, &averageCompletionTime)

		gotStats, found, version, err := repository.GetStats(ctx, filter)
		if err != nil {
			t.Fatalf("get missing stats: %v", err)
		}
		if found || version != 0 || gotStats.TasksCreated != 0 {
			t.Fatalf("expected stats cache miss version=0, got found=%v version=%d stats=%+v", found, version, gotStats)
		}

		if err := repository.SetStats(ctx, filter, version, stats); err != nil {
			t.Fatalf("set stats: %v", err)
		}

		gotStats, found, version, err = repository.GetStats(ctx, filter)
		if err != nil {
			t.Fatalf("get cached stats: %v", err)
		}
		if !found || version != 0 {
			t.Fatalf("expected stats cache hit version=0, got found=%v version=%d", found, version)
		}
		if gotStats.TasksCreated != stats.TasksCreated || gotStats.TasksCompleted != stats.TasksCompleted {
			t.Fatalf("expected cached stats %+v, got %+v", stats, gotStats)
		}
		if gotStats.TasksCompletedRate == nil || *gotStats.TasksCompletedRate != completedRate {
			t.Fatalf("expected completed rate %v, got %v", completedRate, gotStats.TasksCompletedRate)
		}
		if gotStats.TasksAverageCompletionTime == nil || *gotStats.TasksAverageCompletionTime != averageCompletionTime {
			t.Fatalf("expected average completion time %s, got %v", averageCompletionTime, gotStats.TasksAverageCompletionTime)
		}
	})

	t.Run("invalidate stats bumps user and global versions", func(t *testing.T) {
		integration.FlushRedis(t, pool)

		userID := uuid.New()
		filter := newStatsFilter(userID)
		stats := domain.NewStats(4, 3, nil, nil)

		_, found, version := getStats(t, ctx, repository, filter)
		if found || version != 0 {
			t.Fatalf("expected initial miss version=0, got found=%v version=%d", found, version)
		}

		if err := repository.SetStats(ctx, filter, version, stats); err != nil {
			t.Fatalf("set stats: %v", err)
		}

		if err := repository.InvalidateStats(ctx, userID); err != nil {
			t.Fatalf("invalidate stats: %v", err)
		}

		gotStats, found, version := getStats(t, ctx, repository, filter)
		if found || version != 1 || gotStats.TasksCreated != 0 {
			t.Fatalf("expected user stats cache miss version=1 after invalidation, got found=%v version=%d stats=%+v", found, version, gotStats)
		}

		allFilter := stats_feature.NewGetStatsFilter(nil, nil, nil)
		_, found, version = getStats(t, ctx, repository, allFilter)
		if found || version != 1 {
			t.Fatalf("expected global stats cache miss version=1 after invalidation, got found=%v version=%d", found, version)
		}
	})
}

func getStats(
	t *testing.T,
	ctx context.Context,
	repository *stats_redis_repository.StatsRepository,
	filter stats_feature.GetStatsFilter,
) (domain.Stats, bool, int64) {
	t.Helper()

	stats, found, version, err := repository.GetStats(ctx, filter)
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	return stats, found, version
}

func newStatsFilter(userID uuid.UUID) stats_feature.GetStatsFilter {
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	return stats_feature.NewGetStatsFilter(&userID, &from, &to)
}
