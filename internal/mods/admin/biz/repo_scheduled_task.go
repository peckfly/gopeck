package biz

import "context"

type (
	ScheduledTask struct {
		Task Task `json:"task"`
	}

	ScheduledTaskRepository interface {
		AddTask(ctx context.Context, ttl int64, scheduledTask *ScheduledTask) error
		WatchTask(ctx context.Context, execute func(task *ScheduledTask) bool)
	}
)
