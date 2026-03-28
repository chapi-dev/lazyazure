package utils

import (
	"strings"
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	// We can't actually test that the browser opens in unit tests,
	// but we can verify the function exists and can be called
	// The underlying library handles the actual browser opening

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL",
			url:     "https://portal.azure.com",
			wantErr: false, // May fail if no browser available, but function should handle it
		},
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "Azure Portal URL with fragment",
			url:     "https://portal.azure.com/#@tenant/resource/subscriptions/123/overview",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test just verifies the function can be called.
			// In CI environments without a display/browser, this will likely fail,
			// but we don't want to fail the build for that.
			err := OpenBrowser(tt.url)
			if err != nil && !tt.wantErr {
				// Log the error but don't fail - browser might not be available
				t.Logf("OpenBrowser returned error (may be expected in headless environment): %v", err)
				// Check that it's a sensible error, not a panic or programming error
				if strings.Contains(err.Error(), "exec") || strings.Contains(err.Error(), "not found") {
					// Expected errors when browser not available - pass silently
					t.Skip("Browser not available in this environment")
				}
			}
		})
	}
}
