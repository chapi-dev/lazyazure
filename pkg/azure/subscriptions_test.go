package azure

import (
	"testing"
)

func TestDeref_Nil(t *testing.T) {
	result := deref(nil)
	if result != "" {
		t.Errorf("deref(nil) = %q, want empty string", result)
	}
}

func TestDeref_Valid(t *testing.T) {
	input := "test value"
	result := deref(&input)
	if result != "test value" {
		t.Errorf("deref(&input) = %q, want %q", result, input)
	}
}

func TestDeref_EmptyString(t *testing.T) {
	input := ""
	result := deref(&input)
	if result != "" {
		t.Errorf("deref(&empty) = %q, want empty string", result)
	}
}

func TestDeref_SpecialCharacters(t *testing.T) {
	tests := []string{
		"with spaces",
		"with/slashes",
		"with-dashes",
		"with_underscores",
		"with.dots",
		"CamelCase",
		"UPPERCASE",
		"lowercase",
		"Mixed-Case_123",
		"Unicode: 日本語",
	}

	for _, input := range tests {
		result := deref(&input)
		if result != input {
			t.Errorf("deref(%q) = %q, want %q", input, result, input)
		}
	}
}

func TestDeref_LongString(t *testing.T) {
	// Test with a long string to ensure no buffer issues
	input := "a very long string " + string(make([]byte, 10000))
	result := deref(&input)
	if result != input {
		t.Errorf("deref(long string) length = %d, want %d", len(result), len(input))
	}
}
