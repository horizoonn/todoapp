//go:build integration

package tasks_postgres_repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
	tasks_postgres_repository "github.com/horizoonn/todoapp/internal/features/tasks/repository/postgres"
	users_postgres_repository "github.com/horizoonn/todoapp/internal/features/users/repository/postgres"
	"github.com/horizoonn/todoapp/internal/testsupport/integration"
)

func TestTasksRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool := integration.NewPostgresPool(t)
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)

	t.Run("save and get task", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		author := saveUser(t, ctx, usersRepository, "Alice Johnson")
		description := "Finish math homework"
		task := newActiveTask("Homework", &description, author.ID, baseTime())

		savedTask, err := tasksRepository.SaveTask(ctx, task)
		if err != nil {
			t.Fatalf("save task: %v", err)
		}
		if savedTask.ID != task.ID || savedTask.AuthorUserID != author.ID {
			t.Fatalf("expected saved task %+v, got %+v", task, savedTask)
		}

		gotTask, err := tasksRepository.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("get task: %v", err)
		}
		if gotTask.ID != savedTask.ID || gotTask.Description == nil || *gotTask.Description != description {
			t.Fatalf("expected saved task %+v, got %+v", savedTask, gotTask)
		}
	})

	t.Run("list tasks sorted and filtered by author", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		author := saveUser(t, ctx, usersRepository, "Alice Johnson")
		otherAuthor := saveUser(t, ctx, usersRepository, "Bob Smith")
		createdAt := baseTime()

		task := saveTask(t, ctx, tasksRepository, newActiveTask("Homework", nil, author.ID, createdAt))
		otherTask := saveTask(t, ctx, tasksRepository, newActiveTask("Other task", nil, otherAuthor.ID, createdAt.Add(time.Hour)))

		allTasks, err := tasksRepository.GetTasks(ctx, tasks_feature.NewGetTasksFilter(nil, 10, 0))
		if err != nil {
			t.Fatalf("get all tasks: %v", err)
		}
		if len(allTasks) != 2 {
			t.Fatalf("expected 2 tasks, got %+v", allTasks)
		}
		if allTasks[0].ID != otherTask.ID || allTasks[1].ID != task.ID {
			t.Fatalf("expected tasks sorted by created_at desc, got %+v", allTasks)
		}

		authorTasks, err := tasksRepository.GetTasks(ctx, tasks_feature.NewGetTasksFilter(&author.ID, 10, 0))
		if err != nil {
			t.Fatalf("get author tasks: %v", err)
		}
		if len(authorTasks) != 1 || authorTasks[0].ID != task.ID {
			t.Fatalf("expected only task %s for author %s, got %+v", task.ID, author.ID, authorTasks)
		}
	})

	t.Run("update task increments version", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		author := saveUser(t, ctx, usersRepository, "Alice Johnson")
		createdAt := baseTime()
		task := saveTask(t, ctx, tasksRepository, newActiveTask("Homework", nil, author.ID, createdAt))

		completedAt := createdAt.Add(2 * time.Hour)
		task.Title = "Updated homework"
		task.Completed = true
		task.CompletedAt = &completedAt
		updatedTask, err := tasksRepository.UpdateTask(ctx, task)
		if err != nil {
			t.Fatalf("update task: %v", err)
		}
		if updatedTask.Version != task.Version+1 {
			t.Fatalf("expected version %d, got %d", task.Version+1, updatedTask.Version)
		}
		if !updatedTask.Completed || updatedTask.CompletedAt == nil || !updatedTask.CompletedAt.Equal(completedAt) {
			t.Fatalf("expected completed task, got %+v", updatedTask)
		}
	})

	t.Run("stale update returns conflict", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		author := saveUser(t, ctx, usersRepository, "Alice Johnson")
		task := saveTask(t, ctx, tasksRepository, newActiveTask("Homework", nil, author.ID, baseTime()))

		if _, err := tasksRepository.UpdateTask(ctx, task); err != nil {
			t.Fatalf("update task first time: %v", err)
		}

		_, err := tasksRepository.UpdateTask(ctx, task)
		if !errors.Is(err, core_errors.ErrConflict) {
			t.Fatalf("expected ErrConflict for stale version update, got %v", err)
		}
	})

	t.Run("missing author returns invalid argument", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		task := newActiveTask("Task with missing author", nil, uuid.New(), baseTime())
		_, err := tasksRepository.SaveTask(ctx, task)
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for missing author, got %v", err)
		}
	})

	t.Run("delete task and report not found", func(t *testing.T) {
		integration.TruncateAll(t, pool)

		author := saveUser(t, ctx, usersRepository, "Alice Johnson")
		task := saveTask(t, ctx, tasksRepository, newActiveTask("Homework", nil, author.ID, baseTime()))

		if err := tasksRepository.DeleteTask(ctx, task.ID); err != nil {
			t.Fatalf("delete task: %v", err)
		}

		_, err := tasksRepository.GetTask(ctx, task.ID)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound after delete, got %v", err)
		}

		err = tasksRepository.DeleteTask(ctx, task.ID)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for deleting missing task, got %v", err)
		}
	})
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

func saveTask(
	t *testing.T,
	ctx context.Context,
	repository *tasks_postgres_repository.TasksRepository,
	task domain.Task,
) domain.Task {
	t.Helper()

	savedTask, err := repository.SaveTask(ctx, task)
	if err != nil {
		t.Fatalf("save task %q: %v", task.Title, err)
	}
	return savedTask
}

func newActiveTask(title string, description *string, authorID uuid.UUID, createdAt time.Time) domain.Task {
	return domain.NewTask(
		uuid.New(),
		1,
		title,
		description,
		false,
		createdAt,
		nil,
		authorID,
	)
}

func baseTime() time.Time {
	return time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
}
