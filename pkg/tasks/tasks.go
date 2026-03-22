package tasks

import (
	"context"
	"sync"
)

// Task represents an async operation
type Task struct {
	ctx    context.Context
	cancel context.CancelFunc
	fn     func(ctx context.Context)
}

// TaskManager manages async tasks
type TaskManager struct {
	tasks []*Task
	mu    sync.Mutex
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make([]*Task, 0),
	}
}

// NewTask creates and starts a new task
func (tm *TaskManager) NewTask(fn func(ctx context.Context)) *Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ctx:    ctx,
		cancel: cancel,
		fn:     fn,
	}

	tm.tasks = append(tm.tasks, task)

	go func() {
		fn(ctx)
	}()

	return task
}

// StopAll stops all running tasks
func (tm *TaskManager) StopAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, task := range tm.tasks {
		task.cancel()
	}
	tm.tasks = tm.tasks[:0]
}

// Close is an alias for StopAll for compatibility
func (tm *TaskManager) Close() {
	tm.StopAll()
}
