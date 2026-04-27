package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/horizoonn/todoapp/internal/core/domain"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

func TestTaskValidate(t *testing.T) {
	createdAt := time.Date(2024, 4, 25, 10, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(time.Hour)
	beforeCreatedAt := createdAt.Add(-time.Minute)
	description := "Finish math homework"

	tests := []struct {
		name    string
		task    domain.Task
		wantErr bool
	}{
		{
			name: "valid active task",
			task: domain.NewTask(uuid.New(), 1, "Homework", &description, false, createdAt, nil, uuid.New()),
		},
		{
			name: "valid completed task",
			task: domain.NewTask(uuid.New(), 1, "Homework", &description, true, createdAt, &completedAt, uuid.New()),
		},
		{
			name:    "author is required",
			task:    domain.NewTask(uuid.New(), 1, "Homework", nil, false, createdAt, nil, uuid.Nil),
			wantErr: true,
		},
		{
			name:    "title is required",
			task:    domain.NewTask(uuid.New(), 1, "", nil, false, createdAt, nil, uuid.New()),
			wantErr: true,
		},
		{
			name: "description must not be empty",
			task: func() domain.Task {
				description := ""
				return domain.NewTask(uuid.New(), 1, "Homework", &description, false, createdAt, nil, uuid.New())
			}(),
			wantErr: true,
		},
		{
			name:    "completed task needs completed_at",
			task:    domain.NewTask(uuid.New(), 1, "Homework", nil, true, createdAt, nil, uuid.New()),
			wantErr: true,
		},
		{
			name:    "completed_at must not be before created_at",
			task:    domain.NewTask(uuid.New(), 1, "Homework", nil, true, createdAt, &beforeCreatedAt, uuid.New()),
			wantErr: true,
		},
		{
			name:    "active task must not have completed_at",
			task:    domain.NewTask(uuid.New(), 1, "Homework", nil, false, createdAt, &completedAt, uuid.New()),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestTaskApplyPatch(t *testing.T) {
	createdAt := time.Date(2024, 4, 25, 10, 0, 0, 0, time.UTC)
	description := "Initial description"

	tests := []struct {
		name      string
		patch     domain.TaskPatch
		assert    func(t *testing.T, task domain.Task)
		wantErr   bool
		unchanged bool
	}{
		{
			name: "updates title and description",
			patch: domain.NewTaskPatch(
				nullableValue("Updated title"),
				nullableValue("Updated description"),
				nullableUnset[bool](),
			),
			assert: func(t *testing.T, task domain.Task) {
				t.Helper()
				if task.Title != "Updated title" {
					t.Fatalf("expected title to be patched, got %q", task.Title)
				}
				if task.Description == nil || *task.Description != "Updated description" {
					t.Fatalf("expected description to be patched, got %v", task.Description)
				}
			},
		},
		{
			name: "clears description",
			patch: domain.NewTaskPatch(
				nullableUnset[string](),
				nullableNull[string](),
				nullableUnset[bool](),
			),
			assert: func(t *testing.T, task domain.Task) {
				t.Helper()
				if task.Description != nil {
					t.Fatalf("expected description to be cleared, got %q", *task.Description)
				}
			},
		},
		{
			name: "marks task completed",
			patch: domain.NewTaskPatch(
				nullableUnset[string](),
				nullableUnset[string](),
				nullableValue(true),
			),
			assert: func(t *testing.T, task domain.Task) {
				t.Helper()
				if !task.Completed {
					t.Fatal("expected task to be completed")
				}
				if task.CompletedAt == nil {
					t.Fatal("expected completed_at to be set")
				}
				if task.CompletedAt.Before(task.CreatedAt) {
					t.Fatalf("expected completed_at >= created_at, got %s < %s", task.CompletedAt, task.CreatedAt)
				}
			},
		},
		{
			name: "marks task active and clears completed_at",
			patch: domain.NewTaskPatch(
				nullableUnset[string](),
				nullableUnset[string](),
				nullableValue(false),
			),
			assert: func(t *testing.T, task domain.Task) {
				t.Helper()
				if task.Completed {
					t.Fatal("expected task to be active")
				}
				if task.CompletedAt != nil {
					t.Fatalf("expected completed_at to be cleared, got %s", task.CompletedAt)
				}
			},
		},
		{
			name:      "rejects empty patch",
			patch:     domain.NewTaskPatch(nullableUnset[string](), nullableUnset[string](), nullableUnset[bool]()),
			wantErr:   true,
			unchanged: true,
		},
		{
			name: "rejects null title",
			patch: domain.NewTaskPatch(
				nullableNull[string](),
				nullableUnset[string](),
				nullableUnset[bool](),
			),
			wantErr:   true,
			unchanged: true,
		},
		{
			name: "rejects null completed",
			patch: domain.NewTaskPatch(
				nullableUnset[string](),
				nullableUnset[string](),
				nullableNull[bool](),
			),
			wantErr:   true,
			unchanged: true,
		},
		{
			name: "rejects empty description and keeps original task",
			patch: domain.NewTaskPatch(
				nullableUnset[string](),
				nullableValue(""),
				nullableUnset[bool](),
			),
			wantErr:   true,
			unchanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := domain.NewTask(uuid.New(), 1, "Homework", &description, false, createdAt, nil, uuid.New())
			if tt.name == "marks task active and clears completed_at" {
				completedAt := createdAt.Add(time.Hour)
				task = domain.NewTask(task.ID, task.Version, task.Title, task.Description, true, task.CreatedAt, &completedAt, task.AuthorUserID)
			}
			original := task

			err := task.ApplyPatch(tt.patch)
			if tt.wantErr {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				if tt.unchanged && !sameTask(task, original) {
					t.Fatalf("expected task to stay unchanged, got %+v, want %+v", task, original)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tt.assert != nil {
				tt.assert(t, task)
			}
		})
	}
}

func TestTaskCompletionDuration(t *testing.T) {
	createdAt := time.Date(2024, 4, 25, 10, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(90 * time.Minute)

	tests := []struct {
		name string
		task domain.Task
		want *time.Duration
	}{
		{
			name: "active task has no duration",
			task: domain.NewTask(uuid.New(), 1, "Homework", nil, false, createdAt, nil, uuid.New()),
		},
		{
			name: "completed task without completed_at has no duration",
			task: domain.NewTask(uuid.New(), 1, "Homework", nil, true, createdAt, nil, uuid.New()),
		},
		{
			name: "completed task returns duration",
			task: domain.NewTask(uuid.New(), 1, "Homework", nil, true, createdAt, &completedAt, uuid.New()),
			want: durationPtr(90 * time.Minute),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.CompletionDuration()
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil duration, got %s", got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("expected duration %s, got %v", tt.want, got)
			}
		})
	}
}

func durationPtr(duration time.Duration) *time.Duration {
	return &duration
}

func sameTask(a domain.Task, b domain.Task) bool {
	if a.ID != b.ID ||
		a.Version != b.Version ||
		a.Title != b.Title ||
		a.Completed != b.Completed ||
		!a.CreatedAt.Equal(b.CreatedAt) ||
		a.AuthorUserID != b.AuthorUserID {
		return false
	}
	if !sameStringPtr(a.Description, b.Description) {
		return false
	}
	return sameTimePtr(a.CompletedAt, b.CompletedAt)
}

func sameStringPtr(a *string, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func sameTimePtr(a *time.Time, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}
