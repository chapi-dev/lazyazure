package domain

// Resource represents a generic Azure resource
type Resource struct {
	ID             string
	Name           string
	Type           string
	Location       string
	ResourceGroup  string
	SubscriptionID string
	Tags           map[string]string
	Properties     map[string]interface{} // Additional properties for details view
	CreatedTime    string                 // Resource creation time
	ChangedTime    string                 // Last modified time
}

// DisplayString returns a string representation for the UI
func (r *Resource) DisplayString() string {
	return r.Name
}

// GetID returns the resource ID
func (r *Resource) GetID() string {
	return r.ID
}

// GetType returns the resource type (e.g., Microsoft.Compute/virtualMachines)
func (r *Resource) GetType() string {
	return r.Type
}

// GetShortType returns a shortened type name for display (e.g., virtualMachines)
func (r *Resource) GetShortType() string {
	parts := splitType(r.Type)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return r.Type
}

// splitType splits the resource type string (helper function)
func splitType(typeStr string) []string {
	result := make([]string, 0)
	current := ""
	for _, char := range typeStr {
		if char == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
