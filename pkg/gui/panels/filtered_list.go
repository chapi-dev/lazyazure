package panels

import (
	"sort"
	"strings"
	"sync"
)

// FilteredList is a thread-safe list that can be filtered based on display strings
type FilteredList[T any] struct {
	allItems       []T
	displayStrings []string // Formatted display text (e.g., "name (suffix)")
	indices        []int    // Indices of filtered items (into allItems)
	filterText     string
	isFiltering    bool
	mu             sync.RWMutex
}

// NewFilteredList creates a new filtered list
func NewFilteredList[T any]() *FilteredList[T] {
	return &FilteredList[T]{
		allItems:       make([]T, 0),
		displayStrings: make([]string, 0),
		indices:        make([]int, 0),
	}
}

// SetItems sets the items in the list with their display strings
func (fl *FilteredList[T]) SetItems(items []T, getDisplay func(T) string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	fl.allItems = items
	fl.displayStrings = make([]string, len(items))
	for i, item := range items {
		fl.displayStrings[i] = getDisplay(item)
	}

	// Re-apply current filter if active
	if fl.isFiltering && fl.filterText != "" {
		fl.applyFilter()
	} else {
		// Show all items
		fl.indices = make([]int, len(items))
		for i := range items {
			fl.indices[i] = i
		}
	}
}

// GetItems returns all items (unfiltered)
func (fl *FilteredList[T]) GetItems() []T {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return fl.allItems
}

// Len returns the number of filtered items
func (fl *FilteredList[T]) Len() int {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return len(fl.indices)
}

// Get returns the item at the given filtered index
func (fl *FilteredList[T]) Get(idx int) (T, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	var zero T
	if idx < 0 || idx >= len(fl.indices) {
		return zero, false
	}

	return fl.allItems[fl.indices[idx]], true
}

// GetDisplayString returns the display string at the given filtered index
func (fl *FilteredList[T]) GetDisplayString(idx int) (string, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	if idx < 0 || idx >= len(fl.indices) {
		return "", false
	}

	return fl.displayStrings[fl.indices[idx]], true
}

// GetFilteredDisplayStrings returns all display strings for filtered items
func (fl *FilteredList[T]) GetFilteredDisplayStrings() []string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	result := make([]string, len(fl.indices))
	for i, idx := range fl.indices {
		result[i] = fl.displayStrings[idx]
	}
	return result
}

// SetFilter sets the filter text and updates filtered indices
func (fl *FilteredList[T]) SetFilter(text string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	fl.filterText = strings.ToLower(text)
	fl.isFiltering = text != ""
	fl.applyFilter()
}

// ClearFilter removes the current filter
func (fl *FilteredList[T]) ClearFilter() {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	fl.filterText = ""
	fl.isFiltering = false
	fl.indices = make([]int, len(fl.allItems))
	for i := range fl.allItems {
		fl.indices[i] = i
	}
}

// GetFilterText returns the current filter text
func (fl *FilteredList[T]) GetFilterText() string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return fl.filterText
}

// IsFiltering returns true if a filter is active
func (fl *FilteredList[T]) IsFiltering() bool {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return fl.isFiltering
}

// GetFilterStats returns (showing count, total count)
func (fl *FilteredList[T]) GetFilterStats() (showing, total int) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return len(fl.indices), len(fl.allItems)
}

// applyFilter updates indices based on current filterText
// Must be called with lock held
// Uses a two-tier match: exact substring matches first (sorted by position),
// then fuzzy matches (characters in order but not contiguous, sorted by score).
func (fl *FilteredList[T]) applyFilter() {
	if fl.filterText == "" {
		fl.indices = make([]int, len(fl.allItems))
		for i := range fl.allItems {
			fl.indices[i] = i
		}
		return
	}

	type scored struct {
		index int
		score int // lower is better; 0 = exact, positive = fuzzy score
		exact bool
	}

	var matches []scored
	for i, display := range fl.displayStrings {
		lower := strings.ToLower(display)
		if pos := strings.Index(lower, fl.filterText); pos >= 0 {
			// Exact substring match — score by position (earlier = better)
			matches = append(matches, scored{index: i, score: pos, exact: true})
		} else if s, ok := fuzzyScore(lower, fl.filterText); ok {
			matches = append(matches, scored{index: i, score: s, exact: false})
		}
	}

	// Sort: exact matches first (by position), then fuzzy (by score)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].exact != matches[j].exact {
			return matches[i].exact
		}
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		return matches[i].index < matches[j].index
	})

	fl.indices = make([]int, len(matches))
	for i, m := range matches {
		fl.indices[i] = m.index
	}
}

// fuzzyScore checks if all chars of pattern appear in text in order (case-insensitive).
// Returns a score (lower is better) and whether it matched.
// Score penalizes gaps between matched characters and non-start-of-word matches.
func fuzzyScore(text, pattern string) (int, bool) {
	tRunes := []rune(text)
	pRunes := []rune(pattern)
	if len(pRunes) == 0 {
		return 0, true
	}
	if len(pRunes) > len(tRunes) {
		return 0, false
	}

	score := 0
	pi := 0
	lastMatch := -1
	for ti := 0; ti < len(tRunes) && pi < len(pRunes); ti++ {
		if tRunes[ti] == pRunes[pi] {
			if lastMatch >= 0 {
				gap := ti - lastMatch - 1
				score += gap * gap // penalize gaps quadratically
			}
			// Bonus for start-of-word matches (after space, hyphen, underscore, or at pos 0)
			if ti == 0 || tRunes[ti-1] == ' ' || tRunes[ti-1] == '-' || tRunes[ti-1] == '_' {
				// no penalty
			} else {
				score += 1
			}
			lastMatch = ti
			pi++
		}
	}

	if pi < len(pRunes) {
		return 0, false // not all pattern chars found
	}
	return score, true
}

// MapFilteredToOriginal maps a filtered index back to the original index
func (fl *FilteredList[T]) MapFilteredToOriginal(filteredIdx int) (int, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	if filteredIdx < 0 || filteredIdx >= len(fl.indices) {
		return -1, false
	}
	return fl.indices[filteredIdx], true
}

// FindByOriginalIndex finds the filtered index for a given original index
func (fl *FilteredList[T]) FindByOriginalIndex(originalIdx int) (int, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	for filteredIdx, idx := range fl.indices {
		if idx == originalIdx {
			return filteredIdx, true
		}
	}
	return -1, false
}

// FindIndex finds the filtered index of the first item that matches the given predicate
// Returns the index and true if found, -1 and false otherwise
func (fl *FilteredList[T]) FindIndex(matcher func(T) bool) (int, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	for filteredIdx, originalIdx := range fl.indices {
		if matcher(fl.allItems[originalIdx]) {
			return filteredIdx, true
		}
	}
	return -1, false
}
