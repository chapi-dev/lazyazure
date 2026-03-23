package azure

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// APIVersionCache caches API versions for resource providers
type APIVersionCache struct {
	client         *armresources.ProvidersClient
	cache          map[string][]string // provider -> API versions
	subscriptionID string
}

// NewAPIVersionCache creates a new API version cache client
func NewAPIVersionCache(client *Client, subscriptionID string) (*APIVersionCache, error) {
	providersClient, err := armresources.NewProvidersClient(subscriptionID, client.Credential(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create providers client: %w", err)
	}

	return &APIVersionCache{
		client:         providersClient,
		cache:          make(map[string][]string),
		subscriptionID: subscriptionID,
	}, nil
}

// GetLatestAPIVersion returns the latest API version for a resource type
func (c *APIVersionCache) GetLatestAPIVersion(ctx context.Context, resourceType string) (string, error) {
	// Parse resource type to get provider namespace
	// Format: Microsoft.Provider/resourceType
	parts := strings.Split(resourceType, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid resource type format: %s", resourceType)
	}

	providerNamespace := parts[0]
	typeName := parts[1]

	// Check cache first
	if versions, ok := c.cache[resourceType]; ok && len(versions) > 0 {
		// Return first (latest) version
		return versions[0], nil
	}

	// Fetch from Azure
	resp, err := c.client.Get(ctx, providerNamespace, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get provider %s: %w", providerNamespace, err)
	}

	// Find the resource type and its API versions
	for _, rt := range resp.ResourceTypes {
		if rt.ResourceType != nil && *rt.ResourceType == typeName {
			if len(rt.APIVersions) > 0 {
				// Cache the versions
				versions := make([]string, len(rt.APIVersions))
				for i, v := range rt.APIVersions {
					versions[i] = *v
				}
				c.cache[resourceType] = versions
				// Return first (latest) version
				return versions[0], nil
			}
		}
	}

	return "", fmt.Errorf("no API versions found for resource type %s", resourceType)
}
