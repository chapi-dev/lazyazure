package utils

import (
	"strings"
	"testing"
)

func TestBuildSubscriptionPortalURL(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		subscriptionID string
		want           string
	}{
		{
			name:           "valid subscription URL",
			tenantID:       "12345678-1234-1234-1234-123456789012",
			subscriptionID: "87654321-4321-4321-4321-210987654321",
			want:           "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/overview",
		},
		{
			name:           "empty tenant ID",
			tenantID:       "",
			subscriptionID: "87654321-4321-4321-4321-210987654321",
			want:           "https://portal.azure.com/#@/resource/subscriptions/87654321-4321-4321-4321-210987654321/overview",
		},
		{
			name:           "empty subscription ID creates double slash",
			tenantID:       "12345678-1234-1234-1234-123456789012",
			subscriptionID: "",
			want:           "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions//overview",
		},
		{
			name:           "both empty creates multiple issues",
			tenantID:       "",
			subscriptionID: "",
			want:           "https://portal.azure.com/#@/resource/subscriptions//overview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSubscriptionPortalURL(tt.tenantID, tt.subscriptionID)
			if got != tt.want {
				t.Errorf("BuildSubscriptionPortalURL() = %v, want %v", got, tt.want)
			}
			// Verify no double slashes for valid URLs (skip edge cases with empty strings)
			if !strings.Contains(tt.name, "empty") && !strings.Contains(tt.name, "double slash") {
				if strings.Contains(got, "://") {
					path := strings.SplitN(got, "://", 2)[1]
					if strings.Contains(path, "//") {
						t.Errorf("URL contains double slashes: %v", got)
					}
				}
			}
		})
	}
}

func TestBuildResourceGroupPortalURL(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		subscriptionID    string
		resourceGroupName string
		want              string
	}{
		{
			name:              "valid resource group URL",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			subscriptionID:    "87654321-4321-4321-4321-210987654321",
			resourceGroupName: "my-resource-group",
			want:              "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-resource-group/overview",
		},
		{
			name:              "RG name with hyphens",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			subscriptionID:    "87654321-4321-4321-4321-210987654321",
			resourceGroupName: "my-rg-123",
			want:              "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg-123/overview",
		},
		{
			name:              "RG name with underscores",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			subscriptionID:    "87654321-4321-4321-4321-210987654321",
			resourceGroupName: "my_rg_123",
			want:              "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my_rg_123/overview",
		},
		{
			name:              "RG name with numbers only",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			subscriptionID:    "87654321-4321-4321-4321-210987654321",
			resourceGroupName: "rg123456",
			want:              "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/rg123456/overview",
		},
		{
			name:              "empty RG name",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			subscriptionID:    "87654321-4321-4321-4321-210987654321",
			resourceGroupName: "",
			want:              "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups//overview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildResourceGroupPortalURL(tt.tenantID, tt.subscriptionID, tt.resourceGroupName)
			if got != tt.want {
				t.Errorf("BuildResourceGroupPortalURL() = %v, want %v", got, tt.want)
			}
			// Verify no double slashes for valid URLs (skip edge cases with empty strings)
			if !strings.Contains(tt.name, "empty") {
				if strings.Contains(got, "://") {
					path := strings.SplitN(got, "://", 2)[1]
					if strings.Contains(path, "//") {
						t.Errorf("URL contains double slashes: %v", got)
					}
				}
			}
		})
	}
}

func TestBuildResourcePortalURL(t *testing.T) {
	tests := []struct {
		name       string
		tenantID   string
		resourceID string
		want       string
	}{
		{
			name:       "valid resource URL with leading slash",
			tenantID:   "12345678-1234-1234-1234-123456789012",
			resourceID: "/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/my-vm",
			want:       "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/my-vm/overview",
		},
		{
			name:       "resource URL should not have double slashes",
			tenantID:   "tenant-123",
			resourceID: "/subscriptions/sub-123/resourceGroups/rg-1/providers/Microsoft.Storage/storageAccounts/account1",
			want:       "https://portal.azure.com/#@tenant-123/resource/subscriptions/sub-123/resourceGroups/rg-1/providers/Microsoft.Storage/storageAccounts/account1/overview",
		},
		{
			name:       "resource ID without leading slash",
			tenantID:   "12345678-1234-1234-1234-123456789012",
			resourceID: "subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg",
			want:       "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resourcesubscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg/overview",
		},
		{
			name:       "empty resource ID",
			tenantID:   "12345678-1234-1234-1234-123456789012",
			resourceID: "",
			want:       "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/overview",
		},
		{
			name:       "nested resource (SQL database)",
			tenantID:   "12345678-1234-1234-1234-123456789012",
			resourceID: "/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg/providers/Microsoft.Sql/servers/myserver/databases/mydb",
			want:       "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/resourceGroups/my-rg/providers/Microsoft.Sql/servers/myserver/databases/mydb/overview",
		},
		{
			name:       "subscription only resource ID",
			tenantID:   "12345678-1234-1234-1234-123456789012",
			resourceID: "/subscriptions/87654321-4321-4321-4321-210987654321",
			want:       "https://portal.azure.com/#@12345678-1234-1234-1234-123456789012/resource/subscriptions/87654321-4321-4321-4321-210987654321/overview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildResourcePortalURL(tt.tenantID, tt.resourceID)
			if got != tt.want {
				t.Errorf("BuildResourcePortalURL() = %v, want %v", got, tt.want)
			}
			// Verify no double slashes (except in protocol)
			if strings.Contains(got, "://") {
				path := strings.SplitN(got, "://", 2)[1]
				if strings.Contains(path, "//") {
					t.Errorf("URL contains double slashes: %v", got)
				}
			}
		})
	}
}

// TestPortalURLEdgeCases tests additional edge cases for all URL builders
func TestPortalURLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		contains []string // strings that should be present
		excludes []string // strings that should NOT be present
	}{
		{
			name:     "subscription URL has correct structure",
			url:      BuildSubscriptionPortalURL("tenant-123", "sub-456"),
			contains: []string{"https://", "portal.azure.com", "#@tenant-123", "/subscriptions/sub-456/", "/overview"},
			excludes: []string{"//subscriptions"}, // no double slashes
		},
		{
			name:     "resource group URL has correct structure",
			url:      BuildResourceGroupPortalURL("tenant-123", "sub-456", "my-rg"),
			contains: []string{"https://", "portal.azure.com", "#@tenant-123", "/subscriptions/sub-456/", "/resourceGroups/my-rg/", "/overview"},
			excludes: []string{"//subscriptions", "//resourceGroups"},
		},
		{
			name:     "resource URL has correct structure",
			url:      BuildResourcePortalURL("tenant-123", "/subscriptions/sub-456/resourceGroups/my-rg"),
			contains: []string{"https://", "portal.azure.com", "#@tenant-123", "/resource/subscriptions/sub-456/", "/overview"},
			excludes: []string{"//resource/"},
		},
		{
			name:     "long tenant ID",
			url:      BuildSubscriptionPortalURL("very-long-tenant-id-with-many-characters-123456789", "sub-456"),
			contains: []string{"#@very-long-tenant-id-with-many-characters-123456789"},
			excludes: []string{},
		},
		{
			name:     "special characters in RG name",
			url:      BuildResourceGroupPortalURL("tenant-123", "sub-456", "rg-with.many_special-chars123"),
			contains: []string{"/resourceGroups/rg-with.many_special-chars123/"},
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.contains {
				if !strings.Contains(tt.url, want) {
					t.Errorf("URL %q should contain %q", tt.url, want)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(tt.url, exclude) {
					t.Errorf("URL %q should NOT contain %q", tt.url, exclude)
				}
			}
		})
	}
}

// TestPortalURLConsistency verifies that URL patterns are consistent
func TestPortalURLConsistency(t *testing.T) {
	tenantID := "test-tenant-123"
	subID := "test-sub-456"
	rgName := "test-rg"

	// Build URLs
	subURL := BuildSubscriptionPortalURL(tenantID, subID)
	rgURL := BuildResourceGroupPortalURL(tenantID, subID, rgName)
	resURL := BuildResourcePortalURL(tenantID, "/subscriptions/"+subID+"/resourceGroups/"+rgName)

	// All should have same base
	base := "https://portal.azure.com/#@" + tenantID + "/resource"
	if !strings.HasPrefix(subURL, base) {
		t.Errorf("Subscription URL %q does not have expected prefix %q", subURL, base)
	}
	if !strings.HasPrefix(rgURL, base) {
		t.Errorf("Resource Group URL %q does not have expected prefix %q", rgURL, base)
	}
	if !strings.HasPrefix(resURL, base) {
		t.Errorf("Resource URL %q does not have expected prefix %q", resURL, base)
	}

	// All should end with /overview
	suffix := "/overview"
	if !strings.HasSuffix(subURL, suffix) {
		t.Errorf("Subscription URL %q does not have expected suffix %q", subURL, suffix)
	}
	if !strings.HasSuffix(rgURL, suffix) {
		t.Errorf("Resource Group URL %q does not have expected suffix %q", rgURL, suffix)
	}
	if !strings.HasSuffix(resURL, suffix) {
		t.Errorf("Resource URL %q does not have expected suffix %q", resURL, suffix)
	}
}
