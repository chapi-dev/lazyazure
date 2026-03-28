package utils

import (
	"github.com/pkg/browser"
)

// OpenBrowser opens the given URL in the system's default browser
// This is cross-platform and works on Linux (xdg-open), macOS (open), and Windows (start)
func OpenBrowser(url string) error {
	return browser.OpenURL(url)
}
