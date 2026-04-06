package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab represents a tab in the main panel
type Tab int

const (
	SummaryTab Tab = iota
	JSONTab
)

var tabNames = []string{"Summary", "JSON"}

// MainPanel is the details panel with Summary/JSON tabs
type MainPanel struct {
	viewport    viewport.Model
	tab         Tab
	summaryText string
	jsonText    string
	width       int
	height      int
	active      bool
}

// NewMainPanel creates a new main panel
func NewMainPanel() *MainPanel {
	vp := viewport.New(80, 24)
	return &MainPanel{
		viewport: vp,
		tab:      SummaryTab,
		width:    80,
		height:   24,
	}
}

// SetSize updates the panel dimensions
func (mp *MainPanel) SetSize(width, height int) {
	mp.width = width
	mp.height = height
	// Account for tab bar (1 line) and borders
	mp.viewport.Width = width - 2
	mp.viewport.Height = height - 3
}

// SetActive sets whether this panel is currently focused
func (mp *MainPanel) SetActive(active bool) {
	mp.active = active
}

// SetContent sets the content for both tabs
func (mp *MainPanel) SetContent(summary, jsonText string) {
	mp.summaryText = summary
	mp.jsonText = jsonText
	mp.updateContent()
}

// SetSummary sets only the summary content
func (mp *MainPanel) SetSummary(summary string) {
	mp.summaryText = summary
	mp.updateContent()
}

// SetJSON sets only the JSON content
func (mp *MainPanel) SetJSON(jsonText string) {
	mp.jsonText = jsonText
	mp.updateContent()
}

// updateContent updates the displayed content based on current tab
func (mp *MainPanel) updateContent() {
	switch mp.tab {
	case SummaryTab:
		mp.viewport.SetContent(mp.summaryText)
	case JSONTab:
		mp.viewport.SetContent(mp.jsonText)
	}
}

// NextTab switches to the next tab
func (mp *MainPanel) NextTab() {
	if mp.tab == SummaryTab {
		mp.tab = JSONTab
	} else {
		mp.tab = SummaryTab
	}
	mp.updateContent()
}

// PrevTab switches to the previous tab
func (mp *MainPanel) PrevTab() {
	// Same as NextTab with only 2 tabs
	mp.NextTab()
}

// SetTab sets the active tab directly
func (mp *MainPanel) SetTab(tab Tab) {
	mp.tab = tab
	mp.updateContent()
}

// GetTab returns the current tab
func (mp *MainPanel) GetTab() Tab {
	return mp.tab
}

// ScrollUp scrolls the viewport up
func (mp *MainPanel) ScrollUp() {
	mp.viewport.LineUp(1)
}

// ScrollDown scrolls the viewport down
func (mp *MainPanel) ScrollDown() {
	mp.viewport.LineDown(1)
}

// ScrollPageUp scrolls the viewport up by a page
func (mp *MainPanel) ScrollPageUp() {
	mp.viewport.HalfPageUp()
}

// ScrollPageDown scrolls the viewport down by a page
func (mp *MainPanel) ScrollPageDown() {
	mp.viewport.HalfPageDown()
}

// ScrollTop scrolls to the top
func (mp *MainPanel) ScrollTop() {
	mp.viewport.GotoTop()
}

// ScrollBottom scrolls to the bottom
func (mp *MainPanel) ScrollBottom() {
	mp.viewport.GotoBottom()
}

// GetContentLines returns the current content split into lines
func (mp *MainPanel) GetContentLines() []string {
	var content string
	switch mp.tab {
	case SummaryTab:
		content = mp.summaryText
	case JSONTab:
		content = mp.jsonText
	}
	return strings.Split(content, "\n")
}

// SetHighlightedContent sets the viewport content with search highlights
func (mp *MainPanel) SetHighlightedContent(lines []string) {
	content := strings.Join(lines, "\n")
	mp.viewport.SetContent(content)
}

// GotoLine scrolls to make the specified line visible
func (mp *MainPanel) GotoLine(lineNum int) {
	// viewport.YOffset is the top visible line
	// We want to center the match if possible
	visibleLines := mp.viewport.Height
	halfVisible := visibleLines / 2

	targetY := lineNum - halfVisible
	if targetY < 0 {
		targetY = 0
	}

	mp.viewport.YOffset = targetY
}

// Update handles messages for the main panel
func (mp *MainPanel) Update(msg tea.Msg) (*MainPanel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			mp.ScrollUp()
		case "down", "j":
			mp.ScrollDown()
		case "pgup":
			mp.ScrollPageUp()
		case "pgdown":
			mp.ScrollPageDown()
		case "home", "g":
			mp.ScrollTop()
		case "end", "G":
			mp.ScrollBottom()
		case "[":
			mp.PrevTab()
		case "]":
			mp.NextTab()
		}
	}
	mp.viewport, cmd = mp.viewport.Update(msg)
	return mp, cmd
}

// tabBorderWithBottom creates a custom border for tabs
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

// View renders the main panel with nice tabs
func (mp *MainPanel) View() string {
	// Get the border color based on active state
	borderColor := BorderColorInactive
	if mp.active {
		borderColor = BorderColorActive
	}

	// Create tab styles
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")

	inactiveTabStyle := lipgloss.NewStyle().
		Border(inactiveTabBorder).
		BorderForeground(borderColor).
		Padding(0, 1).
		Foreground(WhiteColor)

	activeTabStyle := inactiveTabStyle.
		Border(activeTabBorder).
		Foreground(GreenColor).
		Bold(true)

	// Content window style - no top border to connect with tabs
	windowStyle := lipgloss.NewStyle().
		BorderForeground(borderColor).
		Border(lipgloss.RoundedBorder()).
		UnsetBorderTop().
		Width(mp.width).
		Height(mp.height - 1) // Account for tab row

	// Render tabs
	var renderedTabs []string
	for i, name := range tabNames {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(tabNames)-1, Tab(i) == mp.tab

		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)

		renderedTabs = append(renderedTabs, style.Render(name))
	}

	// Join tabs horizontally
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Get viewport content
	content := mp.viewport.View()

	// Render content window with matching width
	contentWindow := windowStyle.Width(lipgloss.Width(tabsRow)).Render(content)

	// Combine tabs and content
	return lipgloss.JoinVertical(lipgloss.Left, tabsRow, contentWindow)
}
