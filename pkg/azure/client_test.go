package azure

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// This test will fail if not authenticated, but it tests the constructor
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client (expected if not authenticated): %v", err)
	}

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.credential == nil {
		t.Fatal("Client credential is nil")
	}
}

func TestVerifyAuthentication(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	// Test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail if not logged in, but should not hang
	err = client.VerifyAuthentication(ctx)
	if err != nil {
		t.Logf("Authentication verification failed (expected if not logged in): %v", err)
	}
}

func TestVerifyAuthenticationTimeout(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	// Test with very short timeout to ensure it doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- client.VerifyAuthentication(ctx)
	}()

	select {
	case err := <-done:
		t.Logf("Authentication returned (may be error): %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("VerifyAuthentication hung for more than 2 seconds")
	}
}

func TestGetUserInfo(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.GetUserInfo(ctx)
	if err != nil {
		t.Skipf("Could not get user info (expected if not logged in): %v", err)
	}

	if user == nil {
		t.Fatal("GetUserInfo returned nil user")
	}

	// Verify fields are populated (don't log sensitive values)
	if user.Type == "" {
		t.Error("User type is empty")
	}

	if user.UserPrincipalName == "" {
		t.Error("User principal name is empty")
	}

	if user.TenantID == "" {
		t.Error("Tenant ID is empty")
	}

	// Display name should be populated (may be same as UPN for some users)
	if user.DisplayName == "" {
		t.Error("Display name is empty")
	}

	// Log non-sensitive info only
	t.Logf("User info retrieved successfully - Type: %q, HasDisplayName: %v",
		user.Type, user.DisplayName != "")
}
