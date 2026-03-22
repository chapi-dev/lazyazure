package domain

// Subscription represents an Azure subscription
type Subscription struct {
	ID       string
	Name     string
	State    string
	TenantID string
}

// DisplayString returns a string representation for the UI
func (s *Subscription) DisplayString() string {
	return s.Name
}

// GetID returns the subscription ID
func (s *Subscription) GetID() string {
	return s.ID
}
