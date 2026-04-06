package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar displays context-aware help at the bottom of the screen
type StatusBar struct {
	text   string
	width  int
	active bool
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		text:   "Welcome to LazyAzure - Press ? for help",
		width:  80,
		active: false,
	}
}

// SetSize updates the status bar width
func (sb *StatusBar) SetSize(width int) {
	sb.width = width
}

// SetText updates the status bar text
func (sb *StatusBar) SetText(text string) {
	sb.text = text
}

// SetActive sets whether the status bar is "active" (used for visual feedback)
func (sb *StatusBar) SetActive(active bool) {
	sb.active = active
}

// View renders the status bar
func (sb *StatusBar) View() string {
	styles := NewStyles()

	// Pad or truncate text to fit width
	text := sb.text
	renderedWidth := lipgloss.Width(text)

	if renderedWidth < sb.width {
		// Pad with spaces
		text = text + strings.Repeat(" ", sb.width-renderedWidth)
	} else if renderedWidth > sb.width {
		// Truncate with ellipsis
		text = text[:sb.width-3] + "..."
	}

	return styles.StatusBar.Width(sb.width).Render(text)
}

// FormatContextHelp formats help text based on current context
// This is a placeholder - full implementation will depend on the specific UI context
func FormatContextHelp(activePanel string, isFiltering bool) string {
	var parts []string

	parts = append(parts, "Tab/Shift+Tab: switch panels")
	parts = append(parts, "↑/↓/j/k: navigate")
	parts = append(parts, "Enter: select")

	switch activePanel {
	case "subscriptions", "resourcegroups", "resources":
		parts = append(parts, "/: search")
		parts = append(parts, "r: refresh")
		parts = append(parts, "q: quit")
	case "main":
		parts = append(parts, "[/]: switch tabs")
		parts = append(parts, "c: copy URL")
		parts = append(parts, "o: open portal")
	}

	return strings.Join(parts, " | ")
}

// FormatFilterStatus returns status text when filtering
func FormatFilterStatus(panel string, showing, total int) string {
	return fmt.Sprintf("Filtering %s: showing %d of %d | ESC: clear | Enter: confirm", panel, showing, total)
}
