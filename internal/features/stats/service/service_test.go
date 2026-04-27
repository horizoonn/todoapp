package stats_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
	stats_service "github.com/horizoonn/todoapp/internal/features/stats/service"
	stats_service_mocks "github.com/horizoonn/todoapp/internal/features/stats/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetStatsRejectsInvalidDateRange(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, -1)
	ctrl := gomock.NewController(t)

	repository := stats_service_mocks.NewMockStatsRepository(ctrl)
	cache := stats_service_mocks.NewMockStatsCache(ctrl)
	service := stats_service.NewStatsService(repository, cache)

	_, err := service.GetStats(ctx, stats_feature.NewGetStatsFilter(nil, &from, &to))
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetStatsReturnsCachedStats(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	stats := domain.NewStats(10, 5, floatPtr(50), durationPtr(2*time.Hour))
	filter := stats_feature.NewGetStatsFilter(&userID, nil, nil)
	ctrl := gomock.NewController(t)

	repository := stats_service_mocks.NewMockStatsRepository(ctrl)
	cache := stats_service_mocks.NewMockStatsCache(ctrl)
	cache.EXPECT().
		GetStats(ctx, filter).
		Return(stats, true, int64(3), nil)

	service := stats_service.NewStatsService(repository, cache)

	got, err := service.GetStats(ctx, filter)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.TasksCreated != stats.TasksCreated || got.TasksCompleted != stats.TasksCompleted {
		t.Fatalf("expected cached stats %+v, got %+v", stats, got)
	}
}

func TestGetStatsLoadsRepositoryAndStoresCacheOnMiss(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	wantToExclusive := to.AddDate(0, 0, 1)
	filter := stats_feature.NewGetStatsFilter(nil, &from, &to)
	normalizedFilter := stats_feature.NewGetStatsFilter(nil, &from, &wantToExclusive)
	stats := domain.NewStats(10, 5, floatPtr(50), durationPtr(2*time.Hour))
	ctrl := gomock.NewController(t)

	repository := stats_service_mocks.NewMockStatsRepository(ctrl)
	cache := stats_service_mocks.NewMockStatsCache(ctrl)
	gomock.InOrder(
		cache.EXPECT().
			GetStats(ctx, normalizedFilter).
			Return(domain.Stats{}, false, int64(9), nil),
		repository.EXPECT().
			GetStats(ctx, normalizedFilter).
			Return(stats, nil),
		cache.EXPECT().
			SetStats(ctx, normalizedFilter, int64(9), stats).
			Return(nil),
	)

	service := stats_service.NewStatsService(repository, cache)

	got, err := service.GetStats(ctx, filter)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.TasksCreated != stats.TasksCreated || got.TasksCompleted != stats.TasksCompleted {
		t.Fatalf("expected repository stats %+v, got %+v", stats, got)
	}
}

func TestGetStatsSkipsCacheSetWhenCacheReadFails(t *testing.T) {
	ctx := context.Background()
	filter := stats_feature.NewGetStatsFilter(nil, nil, nil)
	stats := domain.NewStats(10, 5, floatPtr(50), nil)
	ctrl := gomock.NewController(t)

	repository := stats_service_mocks.NewMockStatsRepository(ctrl)
	cache := stats_service_mocks.NewMockStatsCache(ctrl)
	gomock.InOrder(
		cache.EXPECT().
			GetStats(ctx, filter).
			Return(domain.Stats{}, false, int64(0), errors.New("redis unavailable")),
		repository.EXPECT().
			GetStats(ctx, filter).
			Return(stats, nil),
	)

	service := stats_service.NewStatsService(repository, cache)

	got, err := service.GetStats(ctx, filter)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.TasksCreated != stats.TasksCreated {
		t.Fatalf("expected repository stats %+v, got %+v", stats, got)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func durationPtr(value time.Duration) *time.Duration {
	return &value
}
