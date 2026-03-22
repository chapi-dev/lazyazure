package panels

import (
	"sync"
)

// FilteredList is a thread-safe list that can be filtered
type FilteredList[T comparable] struct {
	allItems []T
	indices  []int
	mu       sync.RWMutex
}

// NewFilteredList creates a new filtered list
func NewFilteredList[T comparable]() *FilteredList[T] {
	return &FilteredList[T]{
		allItems: make([]T, 0),
		indices:  make([]int, 0),
	}
}

// SetItems sets the items in the list
func (fl *FilteredList[T]) SetItems(items []T) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	fl.allItems = items
	fl.indices = make([]int, len(items))
	for i := range items {
		fl.indices[i] = i
	}
}

// GetItems returns all items
func (fl *FilteredList[T]) GetItems() []T {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return fl.allItems
}

// Len returns the number of items
func (fl *FilteredList[T]) Len() int {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	return len(fl.indices)
}

// Get returns the item at the given index
func (fl *FilteredList[T]) Get(idx int) (T, bool) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	var zero T
	if idx < 0 || idx >= len(fl.indices) {
		return zero, false
	}

	return fl.allItems[fl.indices[idx]], true
}

// GetDisplayStrings returns display strings for all items
func (fl *FilteredList[T]) GetDisplayStrings(getDisplay func(T) string) []string {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	result := make([]string, len(fl.indices))
	for i, idx := range fl.indices {
		result[i] = getDisplay(fl.allItems[idx])
	}
	return result
}
