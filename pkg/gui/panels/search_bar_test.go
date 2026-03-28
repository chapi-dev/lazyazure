package panels

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

// mockGui is a minimal mock for testing
type mockGui struct {
	views map[string]*gocui.View
}

func TestNewSearchBar(t *testing.T) {
	// We can't easily test with real gocui, but we can test the struct creation
	sb := &SearchBar{
		isActive:   false,
		searchText: "",
	}

	if sb.IsActive() {
		t.Error("New search bar should not be active")
	}

	if sb.GetText() != "" {
		t.Errorf("Expected empty text, got '%s'", sb.GetText())
	}
}

func TestSearchBarSetText(t *testing.T) {
	sb := &SearchBar{
		searchText: "",
	}

	// Track callback invocations
	var callbackText string
	sb.onSearch = func(text string) {
		callbackText = text
	}

	sb.SetText("hello")

	if sb.GetText() != "hello" {
		t.Errorf("Expected text 'hello', got '%s'", sb.GetText())
	}

	if callbackText != "hello" {
		t.Errorf("Callback should have been called with 'hello', got '%s'", callbackText)
	}
}

func TestSearchBarHandleRune(t *testing.T) {
	sb := &SearchBar{
		searchText: "",
	}

	var callbackCount int
	sb.onSearch = func(text string) {
		callbackCount++
	}

	// Test valid characters
	sb.HandleRune('a')
	if sb.GetText() != "a" {
		t.Errorf("Expected 'a', got '%s'", sb.GetText())
	}

	sb.HandleRune('B')
	if sb.GetText() != "aB" {
		t.Errorf("Expected 'aB', got '%s'", sb.GetText())
	}

	sb.HandleRune('1')
	if sb.GetText() != "aB1" {
		t.Errorf("Expected 'aB1', got '%s'", sb.GetText())
	}

	sb.HandleRune('-')
	if sb.GetText() != "aB1-" {
		t.Errorf("Expected 'aB1-', got '%s'", sb.GetText())
	}

	if callbackCount != 4 {
		t.Errorf("Expected 4 callbacks, got %d", callbackCount)
	}
}

func TestSearchBarHandleRuneInvalid(t *testing.T) {
	sb := &SearchBar{
		searchText: "test",
	}

	var callbackCount int
	sb.onSearch = func(text string) {
		callbackCount++
	}

	// Test invalid characters (control characters)
	sb.HandleRune('\n')
	sb.HandleRune('\t')
	sb.HandleRune('\x01')
	sb.HandleRune('\x7F')

	if sb.GetText() != "test" {
		t.Errorf("Text should not change for invalid characters, got '%s'", sb.GetText())
	}

	if callbackCount != 0 {
		t.Errorf("Callback should not be called for invalid characters, got %d", callbackCount)
	}
}

func TestSearchBarBackspace(t *testing.T) {
	sb := &SearchBar{
		searchText: "hello",
	}

	var callbackTexts []string
	sb.onSearch = func(text string) {
		callbackTexts = append(callbackTexts, text)
	}

	sb.Backspace()
	if sb.GetText() != "hell" {
		t.Errorf("Expected 'hell' after backspace, got '%s'", sb.GetText())
	}

	sb.Backspace()
	if sb.GetText() != "hel" {
		t.Errorf("Expected 'hel' after second backspace, got '%s'", sb.GetText())
	}

	// Backspace on empty should be safe
	sb.searchText = ""
	sb.Backspace()
	if sb.GetText() != "" {
		t.Errorf("Empty text should stay empty after backspace, got '%s'", sb.GetText())
	}

	if len(callbackTexts) != 2 {
		t.Errorf("Expected 2 callbacks, got %d", len(callbackTexts))
	}
}

func TestSearchBarClear(t *testing.T) {
	sb := &SearchBar{
		searchText: "search text",
	}

	var callbackText string
	sb.onSearch = func(text string) {
		callbackText = text
	}

	sb.Clear()

	if sb.GetText() != "" {
		t.Errorf("Expected empty text after clear, got '%s'", sb.GetText())
	}

	if callbackText != "" {
		t.Errorf("Callback should be called with empty string, got '%s'", callbackText)
	}
}

func TestSearchBarDeleteWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single word", "hello", ""},
		{"multiple words", "hello world", "hello "},
		{"with special chars", "my-resource-name", "my-resource-"},
		{"empty", "", ""},
		{"trailing spaces", "hello   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &SearchBar{
				searchText: tt.input,
			}

			var callbackText string
			sb.onSearch = func(text string) {
				callbackText = text
			}

			sb.DeleteWord()

			if sb.GetText() != tt.expected {
				t.Errorf("DeleteWord() = '%s', expected '%s'", sb.GetText(), tt.expected)
			}

			if callbackText != tt.expected {
				t.Errorf("Callback text = '%s', expected '%s'", callbackText, tt.expected)
			}
		})
	}
}

func TestSearchBarConfirm(t *testing.T) {
	sb := &SearchBar{
		searchText: "test",
	}

	var confirmed bool
	sb.onConfirm = func() {
		confirmed = true
	}

	sb.Confirm()

	if !confirmed {
		t.Error("Confirm callback should have been called")
	}
}

func TestSearchBarCancel(t *testing.T) {
	sb := &SearchBar{
		searchText: "test search",
	}

	var cancelled bool
	sb.onCancel = func() {
		cancelled = true
	}

	sb.Cancel()

	if !cancelled {
		t.Error("Cancel callback should have been called")
	}

	if sb.GetText() != "" {
		t.Errorf("Text should be cleared after cancel, got '%s'", sb.GetText())
	}
}

func TestSearchBarConcurrency(t *testing.T) {
	sb := &SearchBar{
		searchText: "",
	}

	// Set up callback that does some work
	sb.onSearch = func(text string) {
		// Simulate some work
		_ = len(text)
	}

	// Run concurrent operations
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			sb.HandleRune('a')
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			sb.Backspace()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 20; i++ {
			sb.GetText()
			sb.IsActive()
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestSearchBarIsActive(t *testing.T) {
	sb := &SearchBar{
		isActive: false,
	}

	if sb.IsActive() {
		t.Error("Should not be active initially")
	}

	sb.isActive = true

	if !sb.IsActive() {
		t.Error("Should be active after setting")
	}
}
