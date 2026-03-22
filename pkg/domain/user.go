package domain

// User represents the authenticated Azure user
type User struct {
	Name     string
	Email    string
	TenantID string
}

// IsAuthenticated returns true if the user has valid authentication
type IsAuthenticated bool
