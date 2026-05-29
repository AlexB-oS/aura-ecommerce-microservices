// Package auth holds the cross-service security primitives: password hashing
// and JWT issue/verify. Keeping these here guarantees every service hashes and
// validates tokens identically.
package auth

import "golang.org/x/crypto/bcrypt"

// DefaultCost is bcrypt's work factor. 12 is a good balance for 2024-era
// hardware: noticeably slow for an attacker, still a few ms per login.
const DefaultCost = 12

// HashPassword returns a bcrypt hash of the plaintext password. The salt is
// generated and stored inside the returned hash, so no separate salt column is
// needed.
func HashPassword(plaintext string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword reports whether plaintext matches the stored bcrypt hash. It is
// constant-time with respect to the hash, mitigating timing attacks.
func CheckPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
