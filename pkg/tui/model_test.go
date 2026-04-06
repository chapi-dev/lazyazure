package tui

import (
	"strings"
	"testing"

	"github.com/matsest/lazyazure/pkg/gui"
)

func TestModelView(t *testing.T) {
	// Create a model
	versionInfo := gui.VersionInfo{
		Version: "test",
		Commit:  "abc123",
		Date:    "2026-01-01",
	}

	m := NewModel(nil, nil, versionInfo, true)
	m.width = 120
	m.height = 40
	m.calculateLayout()

	// Get the view output
	output := m.View()

	// Check that all expected elements are present
	tests := []struct {
		name     string
		expected string
	}{
		{"Auth panel", "Auth"},
		{"Subscriptions panel", "Subscriptions"},
		{"Resource Groups panel", "Resource Groups"},
		{"Resources panel", "Resources"},
		{"Status bar", "Welcome to LazyAzure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.expected) {
				t.Errorf("View() missing expected content %q\nOutput:\n%s", tt.expected, output)
			}
		})
	}
}
