package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	core_errors "github.com/horizoonn/todoapp/internal/core/errors"
)

const (
	MinTaskTitleLen = 1
	MaxTaskTitleLen = 100

	MinTaskDescriptionLen = 1
	MaxTaskDescriptionLen = 1000
)

type Task struct {
	ID      uuid.UUID
	Version int

	Title       string
	Description *string
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time

	AuthorUserID uuid.UUID
}

func NewTask(
	id uuid.UUID,
	version int,
	title string,
	description *string,
	completed bool,
	createdAt time.Time,
	completedAt *time.Time,
	authorUserID uuid.UUID,
) Task {
	return Task{
		ID:           id,
		Version:      version,
		Title:        title,
		Description:  description,
		Completed:    completed,
		CreatedAt:    createdAt,
		CompletedAt:  completedAt,
		AuthorUserID: authorUserID,
	}
}

func CreateTask(
	title string,
	description *string,
	authorUserID uuid.UUID,
) Task {
	var (
		id                     = uuid.New()
		version                = 1
		completed              = false
		createdAt              = time.Now()
		completedAt *time.Time = nil
	)

	return NewTask(
		id,
		version,
		title,
		description,
		completed,
		createdAt,
		completedAt,
		authorUserID,
	)
}

func (t *Task) CompletionDuration() *time.Duration {
	if !t.Completed {
		return nil
	}

	if t.CompletedAt == nil {
		return nil
	}

	duration := t.CompletedAt.Sub(t.CreatedAt)

	return &duration
}

func (t *Task) Validate() error {
	if t.AuthorUserID == uuid.Nil {
		return fmt.Errorf("`AuthorUserID` can't be empty: %w", core_errors.ErrInvalidArgument)
	}

	titleLen := len([]rune(t.Title))
	if titleLen < MinTaskTitleLen || titleLen > MaxTaskTitleLen {
		return fmt.Errorf("invalid `Title` len: %d (min: %d, max: %d): %w", titleLen, MinTaskTitleLen, MaxTaskTitleLen, core_errors.ErrInvalidArgument)
	}

	if t.Description != nil {
		descriptionLen := len([]rune(*t.Description))
		if descriptionLen < MinTaskDescriptionLen || descriptionLen > MaxTaskDescriptionLen {
			return fmt.Errorf("invalid `Description` len: %d (min: %d, max: %d): %w", descriptionLen, MinTaskDescriptionLen, MaxTaskDescriptionLen, core_errors.ErrInvalidArgument)
		}
	}

	if t.Completed {
		if t.CompletedAt == nil {
			return fmt.Errorf("`CompletedAt` can't be nil if `Completed`==`true`: %w", core_errors.ErrInvalidArgument)
		}

		if t.CompletedAt.Before(t.CreatedAt) {
			return fmt.Errorf("`CompletedAt` can't be before `CreatedAt`: %w", core_errors.ErrInvalidArgument)
		}
	} else {
		if t.CompletedAt != nil {
			return fmt.Errorf("`CompletedAt` must be `nil` if `Completed`==`false`: %w", core_errors.ErrInvalidArgument)
		}
	}

	return nil
}

type TaskPatch struct {
	Title       Nullable[string]
	Description Nullable[string]
	Completed   Nullable[bool]
}

func NewTaskPatch(title Nullable[string], description Nullable[string], completed Nullable[bool]) TaskPatch {
	return TaskPatch{
		Title:       title,
		Description: description,
		Completed:   completed,
	}
}

func (p *TaskPatch) Validate() error {
	if !p.Title.Set && !p.Description.Set && !p.Completed.Set {
		return fmt.Errorf("task patch must contain at least one field: %w", core_errors.ErrInvalidArgument)
	}

	if p.Title.Set && p.Title.Value == nil {
		return fmt.Errorf("`Title` can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	if p.Completed.Set && p.Completed.Value == nil {
		return fmt.Errorf("`Completed` can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (t *Task) ApplyPatch(patch TaskPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate task patch: %w", err)
	}

	tmp := *t

	if patch.Title.Set {
		tmp.Title = *patch.Title.Value
	}

	if patch.Description.Set {
		tmp.Description = patch.Description.Value
	}

	if patch.Completed.Set {
		wasCompleted := tmp.Completed
		tmp.Completed = *patch.Completed.Value

		if tmp.Completed && !wasCompleted {
			completedAt := time.Now()
			tmp.CompletedAt = &completedAt
		} else if !tmp.Completed {
			tmp.CompletedAt = nil
		}
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched task: %w", err)
	}

	*t = tmp

	return nil
}
