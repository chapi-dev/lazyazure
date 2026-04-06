package components

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// Color definitions matching current gocui theme
var (
	// Primary colors
	GreenColor  = lipgloss.Color("120") // ANSI 256 color 120 (bright green)
	WhiteColor  = lipgloss.Color("255") // ANSI 256 color 255 (white)
	GrayColor   = lipgloss.Color("245") // ANSI 256 color 245 (gray)
	BlueColor   = lipgloss.Color("39")  // ANSI 256 color 39 (blue)
	YellowColor = lipgloss.Color("226") // ANSI 256 color 226 (yellow)
	RedColor    = lipgloss.Color("196") // ANSI 256 color 196 (red)
	BlackColor  = lipgloss.Color("0")   // ANSI 256 color 0 (black)

	// UI colors
	BorderColorActive   = GreenColor
	BorderColorInactive = WhiteColor
	TitleColor          = WhiteColor
	SelectedColor       = BlueColor
)

// Base styles
type Styles struct {
	// Panel styles
	ActivePanel   lipgloss.Style
	InactivePanel lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemGray     lipgloss.Style

	// Status bar styles
	StatusBar       lipgloss.Style
	StatusBarActive lipgloss.Style

	// Main panel styles
	MainPanel       lipgloss.Style
	MainPanelTab    lipgloss.Style
	MainPanelTabSel lipgloss.Style

	// Auth panel styles
	AuthPanel lipgloss.Style
}

// NewStyles creates the default styles with the given terminal size
func NewStyles() Styles {
	return Styles{
		// Active panel has green border
		ActivePanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColorActive),

		// Inactive panel has white border
		InactivePanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColorInactive),

		// List item styles
		ListItem:         lipgloss.NewStyle(),
		ListItemSelected: lipgloss.NewStyle().Background(SelectedColor).Foreground(WhiteColor),
		ListItemGray:     lipgloss.NewStyle().Foreground(GrayColor),

		// Status bar - no background to match gocui
		StatusBar:       lipgloss.NewStyle(),
		StatusBarActive: lipgloss.NewStyle().Foreground(GreenColor),

		// Main panel
		MainPanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColorInactive),
		MainPanelTab: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(WhiteColor),
		MainPanelTabSel: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(GreenColor).
			Bold(true),

		// Auth panel
		AuthPanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColorInactive),
	}
}

// SetPanelSize applies width and height to a style
func (s *Styles) SetPanelSize(style lipgloss.Style, width, height int) lipgloss.Style {
	return style.Width(width).Height(height)
}

// PanelTitle returns a styled panel title
func PanelTitle(title string) string {
	return lipgloss.NewStyle().
		Foreground(TitleColor).
		Bold(true).
		Render(" " + title + " ")
}

// WithTitle adds a title to a panel
func WithTitle(content string, title string) string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderColorInactive).
		Render("\n" + content)
}

// FormatWithGraySuffix formats text with a gray suffix (mimicking current behavior)
func FormatWithGraySuffix(name, suffix string) string {
	if suffix == "" {
		return name
	}
	suffixStyle := lipgloss.NewStyle().Foreground(GrayColor)
	return name + " " + suffixStyle.Render("("+suffix+")")
}

// EmbedBorderTitle embeds a title into the top border line of a rendered box.
// This creates a gocui-style inline title that sits on the border itself,
// saving one line of vertical space compared to rendering the title inside.
func EmbedBorderTitle(renderedBox string, title string) string {
	lines := []string{}
	currentLine := ""
	for _, r := range renderedBox {
		if r == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return renderedBox
	}

	// Replace the first line (top border) with title-embedded version
	firstLine := lines[0]

	// Strip ANSI codes to analyze the visible border structure
	stripped := stripANSI(firstLine)
	strippedRunes := []rune(stripped)

	if len(strippedRunes) >= 2 {
		// Find positions of border characters in the stripped version
		leftBorderChar := string(strippedRunes[0])
		rightBorderChar := string(strippedRunes[len(strippedRunes)-1])

		// Calculate inner width (between border chars)
		innerWidth := len(strippedRunes) - 2

		// Calculate title dimensions using visible length (stripped of ANSI codes)
		// but insert the raw title (with ANSI codes) into the border
		titleStripped := stripANSI(title)
		titleWithSpacesVisible := " " + titleStripped + " "
		titleWithSpacesRaw := " " + title + " "

		// Truncate if needed (based on visible length)
		visibleRunes := []rune(titleWithSpacesVisible)
		if len(visibleRunes) > innerWidth {
			maxLen := innerWidth - 3 // Space for "..."
			if maxLen < 1 {
				maxLen = 1
			}
			truncated := 0
			currentWidth := 0
			for i := range visibleRunes {
				w := 1 // Simplified width calculation
				if currentWidth+w > maxLen {
					truncated = i
					break
				}
				currentWidth += w
				truncated = i + 1
			}
			// Truncate the raw title to match visible position
			// This is approximate - proper handling would need to account for ANSI codes
			visibleRunes = visibleRunes[:truncated]
			visibleRunes = append(visibleRunes, '.', '.', '.')
			titleWithSpacesVisible = string(visibleRunes)
			// For simplicity in truncation case, use the stripped version
			titleWithSpacesRaw = titleWithSpacesVisible
		}

		// Calculate filler needed based on visible length
		fillerCount := innerWidth - len(visibleRunes)
		if fillerCount < 0 {
			fillerCount = 0
		}
		filler := ""
		for i := 0; i < fillerCount; i++ {
			filler += "─"
		}

		// Build new top line: preserve any leading ANSI codes, then border + styled title + filler + border
		newTopContent := leftBorderChar + titleWithSpacesRaw + filler + rightBorderChar

		// We need to preserve any ANSI sequences at the start of the line
		// Find first visible (non-ANSI) character position
		ansiEnd := 0
		inEscape := false
		for i, r := range firstLine {
			if inEscape {
				if r == 'm' {
					inEscape = false
				}
				ansiEnd = i + 1
			} else if r == '\x1b' {
				inEscape = true
				ansiEnd = i
			} else {
				break
			}
		}

		// Extract leading ANSI codes (colors, etc.)
		leadingANSI := ""
		if ansiEnd > 0 {
			leadingANSI = firstLine[:ansiEnd]
		}

		lines[0] = leadingANSI + newTopContent
	}

	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
