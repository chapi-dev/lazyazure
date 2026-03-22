package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

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

// GetUserInfo retrieves information about the currently authenticated user using Azure CLI
func (c *Client) GetUserInfo(ctx context.Context) (*domain.User, error) {
	// First verify we can get a token
	_, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate - ensure you're logged in with 'az login': %w", err)
	}

	// Use az account show to get user information
	cmd := exec.CommandContext(ctx, "az", "account", "show", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get account info from Azure CLI: %w", err)
	}

	var accountInfo struct {
		User struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"user"`
		TenantID string `json:"tenantId"`
	}

	if err := json.Unmarshal(output, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to parse account info: %w", err)
	}

	user := &domain.User{
		TenantID: accountInfo.TenantID,
	}

	// Map Azure CLI user type to our format
	switch accountInfo.User.Type {
	case "servicePrincipal":
		user.Type = "serviceprincipal"
		user.UserPrincipalName = accountInfo.User.Name // For SPs, this is the appId
		user.DisplayName = accountInfo.User.Name       // For SPs, use appId as display name
	default:
		user.Type = "user"
		user.UserPrincipalName = accountInfo.User.Name // For users, this is the UPN/email

		// Try to get display name from Microsoft Graph via Azure CLI
		// This may fail if user doesn't have Graph permissions, so we fall back to UPN
		user.DisplayName = user.UserPrincipalName
		cmd2 := exec.CommandContext(ctx, "az", "ad", "signed-in-user", "show", "-o", "json")
		output2, err := cmd2.Output()
		if err == nil {
			var userInfo struct {
				DisplayName string `json:"displayName"`
			}
			if err := json.Unmarshal(output2, &userInfo); err == nil && userInfo.DisplayName != "" {
				user.DisplayName = userInfo.DisplayName
			}
		}
	}

	return user, nil
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
