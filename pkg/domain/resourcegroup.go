package domain

// ResourceGroup represents an Azure resource group
type ResourceGroup struct {
	Name              string
	Location          string
	ID                string
	ProvisioningState string
	Tags              map[string]string
	SubscriptionID    string
}

// DisplayString returns a string representation for the UI
func (rg *ResourceGroup) DisplayString() string {
	return rg.Name
}

// GetID returns the resource group ID
func (rg *ResourceGroup) GetID() string {
	return rg.ID
}
