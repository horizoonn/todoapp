package domain

import "time"

type Stats struct {
	TasksCreated               int
	TasksCompleted             int
	TasksCompletedRate         *float64
	TasksAverageCompletionTime *time.Duration
}

func NewStats(
	tasksCreated int,
	tasksCompleted int,
	tasksCompletedRate *float64,
	tasksAverageCompletionTime *time.Duration,
) Stats {
	return Stats{
		TasksCreated:               tasksCreated,
		TasksCompleted:             tasksCompleted,
		TasksCompletedRate:         tasksCompletedRate,
		TasksAverageCompletionTime: tasksAverageCompletionTime,
	}
}

func CalcStats(tasks []Task) Stats {
	if len(tasks) == 0 {
		return NewStats(0, 0, nil, nil)
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

	const percentMultiplier = 100
	tasksCompletedRate := float64(tasksCompleted) / float64(tasksCreated) * percentMultiplier

	var tasksAverageCompletionTime *time.Duration
	if tasksCompleted > 0 {
		avg := totalCompletionDuration / time.Duration(tasksCompleted)
		tasksAverageCompletionTime = &avg
	}

	return NewStats(tasksCreated, tasksCompleted, &tasksCompletedRate, tasksAverageCompletionTime)
}
