//go:build integration

package tasks_redis_repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
	tasks_redis_repository "github.com/horizoonn/todoapp/internal/features/tasks/repository/redis"
	"github.com/horizoonn/todoapp/internal/testsupport/integration"
)

func TestTasksRedisRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool := integration.NewRedisPool(t)
	repository := tasks_redis_repository.NewTasksRepository(pool)

	t.Run("set get and delete task item", func(t *testing.T) {
		integration.FlushRedis(t, pool)

		task := newTask()
		gotTask, found, err := repository.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("get missing task: %v", err)
		}
		if found {
			t.Fatalf("expected missing task cache, got %+v", gotTask)
		}

		if err := repository.SetTask(ctx, task); err != nil {
			t.Fatalf("set task: %v", err)
		}

		gotTask, found, err = repository.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("get cached task: %v", err)
		}
		if !found {
			t.Fatal("expected task cache hit")
		}
		if gotTask.ID != task.ID || gotTask.Description == nil || *gotTask.Description != *task.Description {
			t.Fatalf("expected cached task %+v, got %+v", task, gotTask)
		}

		if err := repository.DeleteTask(ctx, task.ID); err != nil {
			t.Fatalf("delete task: %v", err)
		}

		_, found, err = repository.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("get deleted task: %v", err)
		}
		if found {
			t.Fatal("expected deleted task cache miss")
		}
	})

	t.Run("set get and invalidate task list", func(t *testing.T) {
		integration.FlushRedis(t, pool)

		task := newTask()
		filter := tasks_feature.NewGetTasksFilter(&task.AuthorUserID, 10, 0)

		tasks, found, version, err := repository.GetTasks(ctx, filter)
		if err != nil {
			t.Fatalf("get missing tasks list: %v", err)
		}
		if found || version != 0 || len(tasks) != 0 {
			t.Fatalf("expected empty list cache version=0 miss, got found=%v version=%d tasks=%+v", found, version, tasks)
		}

		if err := repository.SetTasks(ctx, filter, version, []domain.Task{task}); err != nil {
			t.Fatalf("set tasks list: %v", err)
		}

		tasks, found, version = getTasks(t, ctx, repository, filter)
		if !found || version != 0 || len(tasks) != 1 || tasks[0].ID != task.ID {
			t.Fatalf("expected list cache hit version=0 task=%s, got found=%v version=%d tasks=%+v", task.ID, found, version, tasks)
		}

		if err := repository.InvalidateTasks(ctx, task.AuthorUserID); err != nil {
			t.Fatalf("invalidate tasks: %v", err)
		}

		tasks, found, version = getTasks(t, ctx, repository, filter)
		if found || version != 1 || len(tasks) != 0 {
			t.Fatalf("expected user list cache miss version=1 after invalidation, got found=%v version=%d tasks=%+v", found, version, tasks)
		}

		allFilter := tasks_feature.NewGetTasksFilter(nil, 10, 0)
		_, found, version = getTasks(t, ctx, repository, allFilter)
		if found || version != 1 {
			t.Fatalf("expected global list cache miss version=1 after invalidation, got found=%v version=%d", found, version)
		}
	})
}

func getTasks(
	t *testing.T,
	ctx context.Context,
	repository *tasks_redis_repository.TasksRepository,
	filter tasks_feature.GetTasksFilter,
) ([]domain.Task, bool, int64) {
	t.Helper()

	tasks, found, version, err := repository.GetTasks(ctx, filter)
	if err != nil {
		t.Fatalf("get tasks list: %v", err)
	}
	return tasks, found, version
}

func newTask() domain.Task {
	description := "Finish math homework"
	return domain.NewTask(
		uuid.New(),
		1,
		"Homework",
		&description,
		false,
		time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		nil,
		uuid.New(),
	)
}
