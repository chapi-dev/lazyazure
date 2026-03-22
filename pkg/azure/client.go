package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/matsest/lazyazure/pkg/domain"
)

// Client wraps Azure SDK clients and provides high-level operations
type Client struct {
	credential         azcore.TokenCredential
	subscriptionClient *SubscriptionsClient
}

// InitSubscriptionsClient initializes the subscription client after client creation
func (c *Client) InitSubscriptionsClient() (*SubscriptionsClient, error) {
	return NewSubscriptionsClient(c)
}

// NewClient creates a new Azure client using DefaultAzureCredential
func NewClient() (*Client, error) {
	// Use DefaultAzureCredential which automatically:
	// 1. Checks environment variables
	// 2. Checks for managed identity
	// 3. Falls back to Azure CLI credentials (what we want for this app)
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	return &Client{
		credential: credential,
	}, nil
}

// GetUserInfo retrieves information about the currently authenticated user
func (c *Client) GetUserInfo(ctx context.Context) (*domain.User, error) {
	// For MVP, we'll use a simple approach - get a token and extract claims
	// In production, you might want to use Microsoft Graph API

	// Try to get token to verify authentication works
	_, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate - ensure you're logged in with 'az login': %w", err)
	}

	// For now, return a placeholder - we can enhance this later
	// In a real implementation, we'd decode the JWT token or call Graph API
	return &domain.User{
		Name:     "Authenticated User",
		Email:    "",
		TenantID: "",
	}, nil
}

// VerifyAuthentication checks if the client can authenticate
func (c *Client) VerifyAuthentication(ctx context.Context) error {
	_, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	return err
}

// Credential returns the underlying token credential
func (c *Client) Credential() azcore.TokenCredential {
	return c.credential
}
