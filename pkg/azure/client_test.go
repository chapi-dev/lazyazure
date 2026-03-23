package azure

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestParseResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		expected   map[string]string
	}{
		{
			name:       "full resource ID",
			resourceID: "/subscriptions/12345/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			expected: map[string]string{
				"subscription":  "12345",
				"resourceGroup": "myRG",
				"provider":      "Microsoft.Compute",
			},
		},
		{
			name:       "resource group only",
			resourceID: "/subscriptions/12345/resourceGroups/myRG",
			expected: map[string]string{
				"subscription":  "12345",
				"resourceGroup": "myRG",
			},
		},
		{
			name:       "subscription only",
			resourceID: "/subscriptions/12345",
			expected: map[string]string{
				"subscription": "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseResourceID(tt.resourceID)
			for key, expectedValue := range tt.expected {
				if got, ok := result[key]; !ok || got != expectedValue {
					t.Errorf("parseResourceID(%q)[%s] = %q, want %q", tt.resourceID, key, got, expectedValue)
				}
			}
		})
	}
}

func TestParseAzureToken(t *testing.T) {
	// Create a test JWT token
	claims := azureTokenClaims{
		TenantID:          "tenant-123",
		ObjectID:          "user-456",
		UserPrincipalName: "test@example.com",
		Name:              "Test User",
		AppID:             "",
		Azp:               "",
	}

	// Create token using jwt library
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"tid":  claims.TenantID,
		"oid":  claims.ObjectID,
		"upn":  claims.UserPrincipalName,
		"name": claims.Name,
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	result, err := parseAzureToken(tokenString)
	if err != nil {
		t.Fatalf("parseAzureToken failed: %v", err)
	}

	if result.TenantID != claims.TenantID {
		t.Errorf("TenantID = %q, want %q", result.TenantID, claims.TenantID)
	}
	if result.ObjectID != claims.ObjectID {
		t.Errorf("ObjectID = %q, want %q", result.ObjectID, claims.ObjectID)
	}
	if result.UserPrincipalName != claims.UserPrincipalName {
		t.Errorf("UserPrincipalName = %q, want %q", result.UserPrincipalName, claims.UserPrincipalName)
	}
	if result.Name != claims.Name {
		t.Errorf("Name = %q, want %q", result.Name, claims.Name)
	}
}

func TestMapClaimsToAzureClaims(t *testing.T) {
	mapClaims := jwt.MapClaims{
		"tid":                "test-tenant",
		"oid":                "test-object",
		"upn":                "test@upn.com",
		"preferred_username": "preferred@user.com",
		"appid":              "test-app",
		"azp":                "test-azp",
		"name":               "Test Name",
		"idp":                "test-idp",
		"aud":                "test-aud",
		"iss":                "test-iss",
	}

	result := mapClaimsToAzureClaims(mapClaims)

	tests := []struct {
		got      string
		expected string
		field    string
	}{
		{result.TenantID, "test-tenant", "TenantID"},
		{result.ObjectID, "test-object", "ObjectID"},
		{result.UserPrincipalName, "test@upn.com", "UserPrincipalName"},
		{result.PreferredUsername, "preferred@user.com", "PreferredUsername"},
		{result.AppID, "test-app", "AppID"},
		{result.Azp, "test-azp", "Azp"},
		{result.Name, "Test Name", "Name"},
		{result.IdentityProvider, "test-idp", "IdentityProvider"},
		{result.Audience, "test-aud", "Audience"},
		{result.Issuer, "test-iss", "Issuer"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.field, tt.got, tt.expected)
		}
	}
}

func TestMapClaimsToAzureClaimsMissingFields(t *testing.T) {
	mapClaims := jwt.MapClaims{
		"tid": "only-tenant",
	}

	result := mapClaimsToAzureClaims(mapClaims)

	if result.TenantID != "only-tenant" {
		t.Errorf("TenantID = %q, want %q", result.TenantID, "only-tenant")
	}

	// All other fields should be empty
	if result.ObjectID != "" {
		t.Errorf("ObjectID should be empty, got %q", result.ObjectID)
	}
	if result.UserPrincipalName != "" {
		t.Errorf("UserPrincipalName should be empty, got %q", result.UserPrincipalName)
	}
}
