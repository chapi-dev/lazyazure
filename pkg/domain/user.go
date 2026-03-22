package domain

// User represents the authenticated Azure user
type User struct {
	DisplayName       string
	UserPrincipalName string
	Type              string // "user" or "serviceprincipal"
	TenantID          string
}

// IsAuthenticated returns true if the user has valid authentication
type IsAuthenticated bool
