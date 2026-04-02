package azure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matsest/lazyazure/pkg/domain"
)

func TestNewResourceGroupsClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	// Test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.VerifyAuthentication(ctx)
	if err != nil {
		t.Skipf("Not authenticated, skipping: %v", err)
	}

	// Try to create a resource groups client with a fake subscription ID
	rgClient, err := NewResourceGroupsClient(client, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("Failed to create resource groups client: %v", err)
	}

	if rgClient == nil {
		t.Fatal("NewResourceGroupsClient returned nil")
	}

	t.Log("Successfully created ResourceGroupsClient")
}

func TestListResourceGroupsTimeout(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	// Test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.VerifyAuthentication(ctx)
	if err != nil {
		t.Skipf("Not authenticated, skipping: %v", err)
	}

	// Create client
	rgClient, err := NewResourceGroupsClient(client, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("Failed to create resource groups client: %v", err)
	}

	// Try to list with a longer timeout - this should fail quickly with invalid sub ID
	listCtx, listCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer listCancel()

	done := make(chan struct {
		rgs []*domain.ResourceGroup
		err error
	}, 1)

	go func() {
		rgs, err := rgClient.ListResourceGroups(listCtx)
		done <- struct {
			rgs []*domain.ResourceGroup
			err error
		}{rgs, err}
	}()

	select {
	case result := <-done:
		if result.err != nil {
			t.Logf("ListResourceGroups returned error (expected with fake sub ID): %v", result.err)
		} else {
			t.Logf("ListResourceGroups returned %d resource groups", len(result.rgs))
		}
	case <-time.After(15 * time.Second):
		t.Fatal("ListResourceGroups hung for more than 15 seconds")
	}
}

// Unit tests for convertTags helper function

func TestConvertTags_Nil(t *testing.T) {
	result := convertTags(nil)
	if result == nil {
		t.Error("convertTags(nil) should return empty map, not nil")
	}
	if len(result) != 0 {
		t.Errorf("convertTags(nil) returned map with %d elements, want 0", len(result))
	}
}

func TestConvertTags_Empty(t *testing.T) {
	input := make(map[string]*string)
	result := convertTags(input)
	if len(result) != 0 {
		t.Errorf("convertTags(empty) returned map with %d elements, want 0", len(result))
	}
}

func TestConvertTags_Valid(t *testing.T) {
	env := "production"
	owner := "team-alpha"
	costCenter := "12345"

	input := map[string]*string{
		"env":        &env,
		"owner":      &owner,
		"costCenter": &costCenter,
	}

	result := convertTags(input)

	if len(result) != 3 {
		t.Errorf("convertTags returned map with %d elements, want 3", len(result))
	}

	if result["env"] != "production" {
		t.Errorf("result['env'] = %q, want 'production'", result["env"])
	}
	if result["owner"] != "team-alpha" {
		t.Errorf("result['owner'] = %q, want 'team-alpha'", result["owner"])
	}
	if result["costCenter"] != "12345" {
		t.Errorf("result['costCenter'] = %q, want '12345'", result["costCenter"])
	}
}

func TestConvertTags_NilValues(t *testing.T) {
	env := "production"

	input := map[string]*string{
		"env":         &env,
		"empty":       nil,
		"description": nil,
	}

	result := convertTags(input)

	if len(result) != 3 {
		t.Errorf("convertTags returned map with %d elements, want 3", len(result))
	}

	// Non-nil value should be preserved
	if result["env"] != "production" {
		t.Errorf("result['env'] = %q, want 'production'", result["env"])
	}

	// Nil values should be converted to empty strings
	if result["empty"] != "" {
		t.Errorf("result['empty'] = %q, want empty string", result["empty"])
	}
	if result["description"] != "" {
		t.Errorf("result['description'] = %q, want empty string", result["description"])
	}
}

func TestConvertTags_SpecialCharacters(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{"with spaces", "value with spaces"},
		{"with-dashes", "value-with-dashes"},
		{"with_underscores", "value_with_underscores"},
		{"CamelCase", "CamelCaseValue"},
		{"UPPERCASE", "UPPERCASEVALUE"},
		{"lowercase", "lowercasevalue"},
		{"Mixed-Case_123", "Mixed-Case_Value_123"},
		{"Unicode: 日本語", "日本語の値"},
	}

	input := make(map[string]*string)
	expected := make(map[string]string)
	for _, tt := range tests {
		v := tt.value
		input[tt.key] = &v
		expected[tt.key] = tt.value
	}

	result := convertTags(input)

	for key, want := range expected {
		if got := result[key]; got != want {
			t.Errorf("result[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestConvertTags_LargeMap(t *testing.T) {
	// Test with a large map to ensure performance is reasonable
	input := make(map[string]*string)
	expected := make(map[string]string)

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("tag-%d", i)
		value := fmt.Sprintf("value-%d", i)
		input[key] = &value
		expected[key] = value
	}

	result := convertTags(input)

	if len(result) != 1000 {
		t.Errorf("convertTags returned map with %d elements, want 1000", len(result))
	}

	// Verify a few random entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("tag-%d", i*100)
		if result[key] != expected[key] {
			t.Errorf("result[%q] = %q, want %q", key, result[key], expected[key])
		}
	}
}
