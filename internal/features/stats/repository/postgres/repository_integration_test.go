//go:build integration

package stats_postgres_repository_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	stats_feature "github.com/horizoonn/todoapp/internal/features/stats"
	stats_postgres_repository "github.com/horizoonn/todoapp/internal/features/stats/repository/postgres"
	tasks_postgres_repository "github.com/horizoonn/todoapp/internal/features/tasks/repository/postgres"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	"github.com/horizoonn/todoapp/internal/testsupport/integration"
)

func TestStatsRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool := integration.NewPostgresPool(t)
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepository := stats_postgres_repository.NewStatsRepository(pool)

	t.Run("get all stats", func(t *testing.T) {
		integration.TruncateAll(t, pool)
		seedStatsTasks(t, ctx, usersRepository, tasksRepository)

		allStats, err := statsRepository.GetStats(ctx, stats_feature.NewGetStatsFilter(nil, nil, nil))
		if err != nil {
			t.Fatalf("get all stats: %v", err)
		}
		assertStats(t, allStats, 4, 3, 75, 2*time.Hour)
	})

	t.Run("filter stats by user", func(t *testing.T) {
		integration.TruncateAll(t, pool)
		seed := seedStatsTasks(t, ctx, usersRepository, tasksRepository)

		aliceStats, err := statsRepository.GetStats(ctx, stats_feature.NewGetStatsFilter(&seed.aliceID, nil, nil))
		if err != nil {
			t.Fatalf("get alice stats: %v", err)
		}
		assertStats(t, aliceStats, 3, 2, 66.6666666667, 2*time.Hour)
	})

	t.Run("filter stats by date range", func(t *testing.T) {
		integration.TruncateAll(t, pool)
		seed := seedStatsTasks(t, ctx, usersRepository, tasksRepository)

		from := seed.day2
		to := seed.day3
		rangeStats, err := statsRepository.GetStats(ctx, stats_feature.NewGetStatsFilter(nil, &from, &to))
		if err != nil {
			t.Fatalf("get range stats: %v", err)
		}
		assertStats(t, rangeStats, 2, 2, 100, 150*time.Minute)
	})

	t.Run("empty stats return nil rate and average duration", func(t *testing.T) {
		integration.TruncateAll(t, pool)
		seedStatsTasks(t, ctx, usersRepository, tasksRepository)

		emptyFrom := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		emptyStats, err := statsRepository.GetStats(ctx, stats_feature.NewGetStatsFilter(nil, &emptyFrom, nil))
		if err != nil {
			t.Fatalf("get empty stats: %v", err)
		}
		if emptyStats.TasksCreated != 0 || emptyStats.TasksCompleted != 0 {
			t.Fatalf("expected empty stats, got %+v", emptyStats)
		}
		if emptyStats.TasksCompletedRate != nil {
			t.Fatalf("expected nil completed rate for empty stats, got %v", *emptyStats.TasksCompletedRate)
		}
		if emptyStats.TasksAverageCompletionTime != nil {
			t.Fatalf("expected nil average completion time for empty stats, got %s", *emptyStats.TasksAverageCompletionTime)
		}
	})
}

type statsSeed struct {
	aliceID uuid.UUID
	bobID   uuid.UUID
	day1    time.Time
	day2    time.Time
	day3    time.Time
}

func seedStatsTasks(
	t *testing.T,
	ctx context.Context,
	usersRepository *users_postgres_repository.UsersRepository,
	tasksRepository *tasks_postgres_repository.TasksRepository,
) statsSeed {
	t.Helper()

	alice := saveUser(t, ctx, usersRepository, "Alice Johnson")
	bob := saveUser(t, ctx, usersRepository, "Bob Smith")

	day1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
	day3 := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	saveTask(t, ctx, tasksRepository, newCompletedTask("Alice completed fast", alice.ID, day1, day1.Add(1*time.Hour)))
	saveTask(t, ctx, tasksRepository, newCompletedTask("Alice completed slow", alice.ID, day2, day2.Add(3*time.Hour)))
	saveTask(t, ctx, tasksRepository, newActiveTask("Alice active", alice.ID, day3))
	saveTask(t, ctx, tasksRepository, newCompletedTask("Bob completed", bob.ID, day2, day2.Add(2*time.Hour)))

	return statsSeed{
		aliceID: alice.ID,
		bobID:   bob.ID,
		day1:    day1,
		day2:    day2,
		day3:    day3,
	}
}

func saveUser(
	t *testing.T,
	ctx context.Context,
	repository *users_postgres_repository.UsersRepository,
	fullName string,
) domain.User {
	t.Helper()

	user, err := repository.SaveUser(ctx, domain.NewUser(uuid.New(), 1, fullName, nil))
	if err != nil {
		t.Fatalf("save user %q: %v", fullName, err)
	}
	return user
}

func saveTask(t *testing.T, ctx context.Context, repository *tasks_postgres_repository.TasksRepository, task domain.Task) {
	t.Helper()

	if _, err := repository.SaveTask(ctx, task); err != nil {
		t.Fatalf("save task %q: %v", task.Title, err)
	}
}

func newActiveTask(title string, authorID uuid.UUID, createdAt time.Time) domain.Task {
	return domain.NewTask(
		uuid.New(),
		1,
		title,
		nil,
		false,
		createdAt,
		nil,
		authorID,
	)
}

func newCompletedTask(title string, authorID uuid.UUID, createdAt time.Time, completedAt time.Time) domain.Task {
	return domain.NewTask(
		uuid.New(),
		1,
		title,
		nil,
		true,
		createdAt,
		&completedAt,
		authorID,
	)
}

func assertStats(
	t *testing.T,
	stats domain.Stats,
	tasksCreated int,
	tasksCompleted int,
	completedRate float64,
	averageCompletionTime time.Duration,
) {
	t.Helper()

	if stats.TasksCreated != tasksCreated || stats.TasksCompleted != tasksCompleted {
		t.Fatalf("expected created=%d completed=%d, got %+v", tasksCreated, tasksCompleted, stats)
	}
	if stats.TasksCompletedRate == nil {
		t.Fatalf("expected completed rate %.4f, got nil", completedRate)
	}
	if math.Abs(*stats.TasksCompletedRate-completedRate) > 0.0001 {
		t.Fatalf("expected completed rate %.4f, got %.4f", completedRate, *stats.TasksCompletedRate)
	}
	if stats.TasksAverageCompletionTime == nil {
		t.Fatalf("expected average completion time %s, got nil", averageCompletionTime)
	}
	if *stats.TasksAverageCompletionTime != averageCompletionTime {
		t.Fatalf("expected average completion time %s, got %s", averageCompletionTime, *stats.TasksAverageCompletionTime)
	}
}
