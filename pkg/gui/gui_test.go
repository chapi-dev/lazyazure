package gui

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/matsest/lazyazure/pkg/tasks"
)

// MockGui creates a minimal GUI for testing
type MockGui struct {
	g           *gocui.Gui
	taskManager *tasks.TaskManager
	mu          sync.RWMutex
	counter     int
}

func TestAsyncUpdate(t *testing.T) {
	// This test verifies that async updates don't hang
	gui := &MockGui{
		taskManager: tasks.NewTaskManager(),
	}

	// Simulate async work
	done := make(chan bool, 1)

	gui.taskManager.NewTask(func(ctx context.Context) {
		// Simulate some work
		time.Sleep(100 * time.Millisecond)

		gui.mu.Lock()
		gui.counter++
		gui.mu.Unlock()

		done <- true
	})

	select {
	case <-done:
		t.Log("Async task completed successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Async task hung for more than 2 seconds")
	}

	gui.mu.RLock()
	if gui.counter != 1 {
		t.Fatalf("Expected counter to be 1, got %d", gui.counter)
	}
	gui.mu.RUnlock()
}

func TestConcurrentUpdates(t *testing.T) {
	gui := &MockGui{
		taskManager: tasks.NewTaskManager(),
	}

	numTasks := 10
	done := make(chan bool, numTasks)

	for i := 0; i < numTasks; i++ {
		gui.taskManager.NewTask(func(ctx context.Context) {
			gui.mu.Lock()
			gui.counter++
			gui.mu.Unlock()
			done <- true
		})
	}

	completed := 0
	timeout := time.After(3 * time.Second)

	for completed < numTasks {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatalf("Only %d of %d tasks completed before timeout", completed, numTasks)
		}
	}

	gui.mu.RLock()
	if gui.counter != numTasks {
		t.Fatalf("Expected counter to be %d, got %d", numTasks, gui.counter)
	}
	gui.mu.RUnlock()
}
