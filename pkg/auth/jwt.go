package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType distinguishes short-lived access tokens from long-lived refresh
// tokens. Encoding it into the claims stops a refresh token from ever being
// accepted where an access token is required, and vice-versa.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Platform roles carried in the JWT "role" claim. They live here (with the
// claims) so any service can authorize without importing another service's
// domain package.
const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
)

// ErrInvalidToken is returned for any token that fails to parse, has the wrong
// signature, is expired, or has the wrong type.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the JWT payload. It embeds jwt.RegisteredClaims so we get standard
// fields (exp, iat, sub, jti) for free.
type Claims struct {
	Type  TokenType `json:"typ"`
	Email string    `json:"email,omitempty"`
	Role  string    `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// UserID is a convenience accessor for the "sub" claim.
func (c Claims) UserID() string { return c.Subject }

// Manager issues and verifies JWTs with a shared HMAC secret. The same secret
// must be configured on any service that needs to validate tokens (so the API
// gateway or downstream services can authenticate requests without calling the
// user service on every hop).
type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager builds a token manager. accessTTL is typically minutes; refreshTTL
// is typically days.
func NewManager(secret, issuer string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Issue creates a signed token of the given type for a user. The role is
// embedded so any service can make authorization decisions (e.g. admin-only
// endpoints) by validating the token alone, with no call back to the user
// service.
func (m *Manager) Issue(userID, email, role string, typ TokenType) (string, time.Time, error) {
	ttl := m.accessTTL
	if typ == RefreshToken {
		ttl = m.refreshTTL
	}
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := Claims{
		Type:  typ,
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// AccessTTL exposes the configured access-token lifetime so handlers can report
// expires_in to clients.
func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }

// Verify parses and validates a token, additionally checking that its type
// matches `want`. It returns the claims on success.
func (m *Manager) Verify(tokenString string, want TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// Reject any token not signed with HMAC to avoid the classic
		// "alg: none" and RS256/HS256 confusion attacks.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil || !token.Valid || claims.Type != want {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
