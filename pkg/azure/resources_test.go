package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// Tests for parseResourceID

func TestParseResourceID_FullResourceID(t *testing.T) {
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM"
	result := parseResourceID(resourceID)

	if len(result) != 3 {
		t.Errorf("parseResourceID returned %d components, want 3", len(result))
	}

	if result["subscription"] != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("subscription = %q, want %q", result["subscription"], "12345678-1234-1234-1234-123456789012")
	}

	if result["resourceGroup"] != "myRG" {
		t.Errorf("resourceGroup = %q, want %q", result["resourceGroup"], "myRG")
	}

	if result["provider"] != "Microsoft.Compute" {
		t.Errorf("provider = %q, want %q", result["provider"], "Microsoft.Compute")
	}
}

func TestParseResourceID_ResourceGroupOnly(t *testing.T) {
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG"
	result := parseResourceID(resourceID)

	if len(result) != 2 {
		t.Errorf("parseResourceID returned %d components, want 2", len(result))
	}

	if result["subscription"] != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("subscription = %q, want %q", result["subscription"], "12345678-1234-1234-1234-123456789012")
	}

	if result["resourceGroup"] != "myRG" {
		t.Errorf("resourceGroup = %q, want %q", result["resourceGroup"], "myRG")
	}

	if _, ok := result["provider"]; ok {
		t.Error("provider should not be present for resource group only ID")
	}
}

func TestParseResourceID_SubscriptionOnly(t *testing.T) {
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012"
	result := parseResourceID(resourceID)

	if len(result) != 1 {
		t.Errorf("parseResourceID returned %d components, want 1", len(result))
	}

	if result["subscription"] != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("subscription = %q, want %q", result["subscription"], "12345678-1234-1234-1234-123456789012")
	}

	if _, ok := result["resourceGroup"]; ok {
		t.Error("resourceGroup should not be present for subscription only ID")
	}
}

func TestParseResourceID_Empty(t *testing.T) {
	result := parseResourceID("")

	if len(result) != 0 {
		t.Errorf("parseResourceID('') returned %d components, want 0", len(result))
	}
}

func TestParseResourceID_InvalidFormat(t *testing.T) {
	// Various invalid formats
	invalidIDs := []string{
		"not-a-valid-id",
		"subscriptions", // missing leading slash
		"/",             // just slash
		"//",            // double slashes
	}

	for _, id := range invalidIDs {
		result := parseResourceID(id)
		// Should return empty map or partial results without panicking
		if result == nil {
			t.Errorf("parseResourceID(%q) returned nil, want empty map", id)
		}
	}
}

func TestParseResourceID_DoubleSlashes(t *testing.T) {
	// Test that double slashes are handled gracefully
	resourceID := "/subscriptions//12345//resourceGroups//myRG"
	result := parseResourceID(resourceID)

	// Should handle without panic (behavior may vary, but shouldn't crash)
	if result == nil {
		t.Error("parseResourceID with double slashes returned nil")
	}
}

func TestParseResourceID_NoLeadingSlash(t *testing.T) {
	resourceID := "subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG"
	result := parseResourceID(resourceID)

	// Should still work without leading slash
	if result["subscription"] != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("subscription = %q, want %q", result["subscription"], "12345678-1234-1234-1234-123456789012")
	}
}

func TestParseResourceID_NestedResource(t *testing.T) {
	// Test nested resource like SQL database
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Sql/servers/myServer/databases/myDB"
	result := parseResourceID(resourceID)

	if result["subscription"] != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("subscription = %q, want %q", result["subscription"], "12345678-1234-1234-1234-123456789012")
	}

	if result["resourceGroup"] != "myRG" {
		t.Errorf("resourceGroup = %q, want %q", result["resourceGroup"], "myRG")
	}

	if result["provider"] != "Microsoft.Sql" {
		t.Errorf("provider = %q, want %q", result["provider"], "Microsoft.Sql")
	}
}

func TestParseResourceID_SpecialCharacters(t *testing.T) {
	// Resource group and resource names can have various characters
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/my-rg_123/providers/Microsoft.Storage/storageAccounts/myaccount123"
	result := parseResourceID(resourceID)

	if result["resourceGroup"] != "my-rg_123" {
		t.Errorf("resourceGroup = %q, want %q", result["resourceGroup"], "my-rg_123")
	}
}

// Tests for extractPropertiesGeneric

func TestExtractPropertiesGeneric_Nil(t *testing.T) {
	result := extractPropertiesGeneric(nil)
	if result == nil {
		t.Error("extractPropertiesGeneric(nil) should return empty map, not nil")
	}
	if len(result) != 0 {
		t.Errorf("extractPropertiesGeneric(nil) returned map with %d elements, want 0", len(result))
	}
}

func TestExtractPropertiesGeneric_Empty(t *testing.T) {
	input := make(map[string]interface{})
	result := extractPropertiesGeneric(input)
	if len(result) != 0 {
		t.Errorf("extractPropertiesGeneric(empty) returned map with %d elements, want 0", len(result))
	}
}

func TestExtractPropertiesGeneric_ValidMap(t *testing.T) {
	input := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"nested": map[string]interface{}{
			"key": "nestedValue",
		},
	}

	result := extractPropertiesGeneric(input)

	if len(result) != 4 {
		t.Errorf("extractPropertiesGeneric returned map with %d elements, want 4", len(result))
	}

	if result["string"] != "value" {
		t.Errorf("result['string'] = %v, want 'value'", result["string"])
	}
	if result["number"] != float64(42) { // JSON numbers become float64
		t.Errorf("result['number'] = %v, want 42", result["number"])
	}
	if result["bool"] != true {
		t.Errorf("result['bool'] = %v, want true", result["bool"])
	}

	// Check nested map
	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Errorf("result['nested'] is not a map, got %T", result["nested"])
	} else if nested["key"] != "nestedValue" {
		t.Errorf("result['nested']['key'] = %v, want 'nestedValue'", nested["key"])
	}
}

func TestExtractPropertiesGeneric_Struct(t *testing.T) {
	// Test with a struct input
	type TestStruct struct {
		Name    string `json:"name"`
		Value   int    `json:"value"`
		Enabled bool   `json:"enabled"`
	}

	input := TestStruct{
		Name:    "test",
		Value:   100,
		Enabled: true,
	}

	result := extractPropertiesGeneric(input)

	if len(result) != 3 {
		t.Errorf("extractPropertiesGeneric returned map with %d elements, want 3", len(result))
	}

	if result["name"] != "test" {
		t.Errorf("result['name'] = %v, want 'test'", result["name"])
	}
	if result["value"] != float64(100) {
		t.Errorf("result['value'] = %v, want 100", result["value"])
	}
	if result["enabled"] != true {
		t.Errorf("result['enabled'] = %v, want true", result["enabled"])
	}
}

// Tests for extractProperties

func TestExtractProperties_Nil(t *testing.T) {
	result := extractProperties(nil)
	if result == nil {
		t.Error("extractProperties(nil) should return empty map, not nil")
	}
	if len(result) != 0 {
		t.Errorf("extractProperties(nil) returned map with %d elements, want 0", len(result))
	}
}

func TestExtractProperties_ManagedBy(t *testing.T) {
	managedBy := "/subscriptions/123/resourceGroups/otherRG"
	res := &armresources.GenericResourceExpanded{
		ManagedBy: &managedBy,
	}

	result := extractProperties(res)

	if len(result) != 1 {
		t.Errorf("extractProperties returned map with %d elements, want 1", len(result))
	}

	if result["managedBy"] != managedBy {
		t.Errorf("result['managedBy'] = %v, want %v", result["managedBy"], managedBy)
	}
}

func TestExtractProperties_Kind(t *testing.T) {
	kind := "StorageV2"
	res := &armresources.GenericResourceExpanded{
		Kind: &kind,
	}

	result := extractProperties(res)

	if len(result) != 1 {
		t.Errorf("extractProperties returned map with %d elements, want 1", len(result))
	}

	if result["kind"] != kind {
		t.Errorf("result['kind'] = %v, want %v", result["kind"], kind)
	}
}

func TestExtractProperties_MultipleFields(t *testing.T) {
	managedBy := "/subscriptions/123/resourceGroups/otherRG"
	kind := "StorageV2"
	res := &armresources.GenericResourceExpanded{
		ManagedBy: &managedBy,
		Kind:      &kind,
	}

	result := extractProperties(res)

	if len(result) != 2 {
		t.Errorf("extractProperties returned map with %d elements, want 2", len(result))
	}

	if result["managedBy"] != managedBy {
		t.Errorf("result['managedBy'] = %v, want %v", result["managedBy"], managedBy)
	}
	if result["kind"] != kind {
		t.Errorf("result['kind'] = %v, want %v", result["kind"], kind)
	}
}

func TestExtractProperties_NilFields(t *testing.T) {
	// Test when ManagedBy and Kind are nil
	res := &armresources.GenericResourceExpanded{}

	result := extractProperties(res)

	// Should not include nil fields
	if _, ok := result["managedBy"]; ok {
		t.Error("result should not contain 'managedBy' when nil")
	}
	if _, ok := result["kind"]; ok {
		t.Error("result should not contain 'kind' when nil")
	}
}
