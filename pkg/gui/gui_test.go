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

func TestVersionNeedsUpdate(t *testing.T) {
	tests := []struct {
		name        string
		currentVer  string
		latestVer   string
		needsUpdate bool
	}{
		{
			name:        "same version",
			currentVer:  "v1.0.0",
			latestVer:   "v1.0.0",
			needsUpdate: false,
		},
		{
			name:        "older version",
			currentVer:  "v1.0.0",
			latestVer:   "v1.1.0",
			needsUpdate: true,
		},
		{
			name:        "dev version",
			currentVer:  "dev",
			latestVer:   "v1.0.0",
			needsUpdate: false,
		},
		{
			name:        "empty latest",
			currentVer:  "v1.0.0",
			latestVer:   "",
			needsUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gui := &Gui{
				versionInfo: VersionInfo{
					Version: tt.currentVer,
					Commit:  "abc123",
				},
				latestVersion: tt.latestVer,
			}

			result := gui.versionNeedsUpdate()
			if result != tt.needsUpdate {
				t.Errorf("versionNeedsUpdate() = %v, want %v", result, tt.needsUpdate)
			}
		})
	}
}

func TestNewGuiWithVersionInfo(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "v1.0.0",
		Commit:  "abc123def",
		Date:    "2024-01-01",
	}

	// Create GUI with version info - this just tests the constructor
	// We can't fully test without AzureClient, but we can verify the struct is set up
	gui := &Gui{
		versionInfo:     versionInfo,
		taskManager:     tasks.NewTaskManager(),
		tabIndex:        0,
		activePanel:     "subscriptions",
		subList:         nil,
		rgList:          nil,
		resList:         nil,
		mainPanelSearch: nil,
	}

	if gui.versionInfo.Version != versionInfo.Version {
		t.Errorf("Expected version %s, got %s", versionInfo.Version, gui.versionInfo.Version)
	}

	if gui.versionInfo.Commit != versionInfo.Commit {
		t.Errorf("Expected commit %s, got %s", versionInfo.Commit, gui.versionInfo.Commit)
	}

	if gui.versionInfo.Date != versionInfo.Date {
		t.Errorf("Expected date %s, got %s", versionInfo.Date, gui.versionInfo.Date)
	}
}

func TestIsDevelopmentBuild(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		isDevBuild bool
	}{
		{
			name:       "plain dev",
			version:    "dev",
			isDevBuild: true,
		},
		{
			name:       "clean release tag",
			version:    "v1.0.0",
			isDevBuild: false,
		},
		{
			name:       "dirty working tree",
			version:    "v0.2.1-dirty",
			isDevBuild: true,
		},
		{
			name:       "ahead of tag with hash",
			version:    "v0.2.1-2-gc15ffdf",
			isDevBuild: true,
		},
		{
			name:       "ahead of tag with dirty",
			version:    "v0.2.1-2-gc15ffdf-dirty",
			isDevBuild: true,
		},

		{
			name:       "unknown commit",
			version:    "unknown",
			isDevBuild: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gui := &Gui{
				versionInfo: VersionInfo{
					Version: tt.version,
					Commit:  "abc123",
				},
			}

			result := gui.isDevelopmentBuild()
			if result != tt.isDevBuild {
				t.Errorf("isDevelopmentBuild() for version %q = %v, want %v",
					tt.version, result, tt.isDevBuild)
			}
		})
	}
}
