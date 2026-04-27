package tasks_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
	tasks_feature "github.com/horizoonn/todoapp/internal/features/tasks"
	tasks_service "github.com/horizoonn/todoapp/internal/features/tasks/service"
	tasks_service_mocks "github.com/horizoonn/todoapp/internal/features/tasks/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetTaskReturnsCachedTask(t *testing.T) {
	ctx := context.Background()
	task := newTask()
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	cache.EXPECT().
		GetTask(ctx, task.ID).
		Return(task, true, nil)

	service := tasks_service.NewTasksService(repository, cache, nil)

	got, err := service.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ID != task.ID {
		t.Fatalf("expected cached task id %s, got %s", task.ID, got.ID)
	}
}

func TestGetTaskLoadsRepositoryAndStoresCacheOnMiss(t *testing.T) {
	ctx := context.Background()
	task := newTask()
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	gomock.InOrder(
		cache.EXPECT().
			GetTask(ctx, task.ID).
			Return(domain.Task{}, false, nil),
		repository.EXPECT().
			GetTask(ctx, task.ID).
			Return(task, nil),
		cache.EXPECT().
			SetTask(ctx, task).
			Return(nil),
	)

	service := tasks_service.NewTasksService(repository, cache, nil)

	got, err := service.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ID != task.ID {
		t.Fatalf("expected repository task id %s, got %s", task.ID, got.ID)
	}
}

func TestGetTasksReturnsCachedList(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	limit := 10
	offset := 5
	tasks := []domain.Task{newTask()}
	filter := tasks_feature.NewGetTasksFilter(&userID, limit, offset)
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	cache.EXPECT().
		GetTasks(ctx, filter).
		Return(tasks, true, int64(3), nil)

	service := tasks_service.NewTasksService(repository, cache, nil)

	got, err := service.GetTasks(ctx, &userID, &limit, &offset)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != tasks[0].ID {
		t.Fatalf("expected cached tasks %+v, got %+v", tasks, got)
	}
}

func TestGetTasksLoadsRepositoryAndStoresCacheWithVersionOnMiss(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	limit := 10
	offset := 5
	tasks := []domain.Task{newTask()}
	filter := tasks_feature.NewGetTasksFilter(&userID, limit, offset)
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	gomock.InOrder(
		cache.EXPECT().
			GetTasks(ctx, filter).
			Return(nil, false, int64(7), nil),
		repository.EXPECT().
			GetTasks(ctx, filter).
			Return(tasks, nil),
		cache.EXPECT().
			SetTasks(ctx, filter, int64(7), tasks).
			Return(nil),
	)

	service := tasks_service.NewTasksService(repository, cache, nil)

	got, err := service.GetTasks(ctx, &userID, &limit, &offset)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 || got[0].ID != tasks[0].ID {
		t.Fatalf("expected repository tasks %+v, got %+v", tasks, got)
	}
}

func TestCreateTaskSavesTaskAndInvalidatesCaches(t *testing.T) {
	ctx := context.Background()
	authorID := uuid.New()
	description := "Finish math homework"
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	statsCache := tasks_service_mocks.NewMockStatsCacheInvalidator(ctrl)
	gomock.InOrder(
		repository.EXPECT().
			SaveTask(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, task domain.Task) (domain.Task, error) {
				if task.AuthorUserID != authorID {
					t.Fatalf("expected author id %s, got %s", authorID, task.AuthorUserID)
				}
				return task, nil
			}),
		cache.EXPECT().
			SetTask(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, task domain.Task) error {
				if task.AuthorUserID != authorID {
					t.Fatalf("expected cached task author id %s, got %s", authorID, task.AuthorUserID)
				}
				return nil
			}),
		cache.EXPECT().
			InvalidateTasks(ctx, authorID).
			Return(nil),
		statsCache.EXPECT().
			InvalidateStats(ctx, authorID).
			Return(nil),
	)

	service := tasks_service.NewTasksService(repository, cache, statsCache)

	task, err := service.CreateTask(ctx, "Homework", &description, authorID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if task.AuthorUserID != authorID {
		t.Fatalf("expected author id %s, got %s", authorID, task.AuthorUserID)
	}
}

func TestCreateTaskValidationFailureDoesNotSave(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	statsCache := tasks_service_mocks.NewMockStatsCacheInvalidator(ctrl)
	service := tasks_service.NewTasksService(repository, cache, statsCache)

	_, err := service.CreateTask(ctx, "", nil, uuid.New())
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDeleteTaskDeletesTaskAndInvalidatesCaches(t *testing.T) {
	ctx := context.Background()
	task := newTask()
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	statsCache := tasks_service_mocks.NewMockStatsCacheInvalidator(ctrl)
	gomock.InOrder(
		repository.EXPECT().
			GetTask(ctx, task.ID).
			Return(task, nil),
		repository.EXPECT().
			DeleteTask(ctx, task.ID).
			Return(nil),
		cache.EXPECT().
			DeleteTask(ctx, task.ID).
			Return(nil),
		cache.EXPECT().
			InvalidateTasks(ctx, task.AuthorUserID).
			Return(nil),
		statsCache.EXPECT().
			InvalidateStats(ctx, task.AuthorUserID).
			Return(nil),
	)

	service := tasks_service.NewTasksService(repository, cache, statsCache)

	if err := service.DeleteTask(ctx, task.ID); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestPatchTaskUpdatesTaskAndInvalidatesCaches(t *testing.T) {
	ctx := context.Background()
	task := newTask()
	updatedTitle := "Updated title"
	updatedTask := task
	updatedTask.Title = updatedTitle
	patchedTask := updatedTask
	patchedTask.Version++
	ctrl := gomock.NewController(t)

	repository := tasks_service_mocks.NewMockTasksRepository(ctrl)
	cache := tasks_service_mocks.NewMockTasksCache(ctrl)
	statsCache := tasks_service_mocks.NewMockStatsCacheInvalidator(ctrl)
	gomock.InOrder(
		repository.EXPECT().
			GetTask(ctx, task.ID).
			Return(task, nil),
		repository.EXPECT().
			UpdateTask(ctx, updatedTask).
			Return(patchedTask, nil),
		cache.EXPECT().
			SetTask(ctx, patchedTask).
			Return(nil),
		cache.EXPECT().
			InvalidateTasks(ctx, patchedTask.AuthorUserID).
			Return(nil),
		statsCache.EXPECT().
			InvalidateStats(ctx, patchedTask.AuthorUserID).
			Return(nil),
	)

	service := tasks_service.NewTasksService(repository, cache, statsCache)

	got, err := service.PatchTask(ctx, task.ID, domain.NewTaskPatch(
		domain.Nullable[string]{Value: &updatedTitle, Set: true},
		domain.Nullable[string]{},
		domain.Nullable[bool]{},
	))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != updatedTitle {
		t.Fatalf("expected patched title %q, got %q", updatedTitle, got.Title)
	}
}

func newTask() domain.Task {
	description := "Finish math homework"
	return domain.NewTask(
		uuid.New(),
		1,
		"Homework",
		&description,
		false,
		time.Date(2024, 4, 25, 10, 0, 0, 0, time.UTC),
		nil,
		uuid.New(),
	)
}
