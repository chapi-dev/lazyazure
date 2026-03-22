package azure

import (
	"context"
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
