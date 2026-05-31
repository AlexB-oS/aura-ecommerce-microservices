// Package domain holds the core business entities and errors for the user
// service. It has no knowledge of HTTP, SQL, or JSON — that isolation is what
// makes the business rules easy to test and the service easy to evolve.
package domain

import (
	"errors"
	"time"
)

// Role enumerates the access levels a user can have. Keeping it here (rather
// than as free-form strings) means new roles are added in one place.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User is the central entity of the service. PasswordHash never leaves the
// service boundary — the transport layer maps User onto a DTO that omits it.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Domain errors are sentinel values the service layer returns and the transport
// layer maps onto HTTP status codes. Comparing with errors.Is keeps the layers
// decoupled from each other's representations.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
