// Package events defines the user service's published event payloads — the
// "contract" other services consume off the NATS bus. Keeping the payload
// shapes in one named place makes the integration surface explicit and easy to
// document.
package events

import pkgevents "github.com/aura-commerce/platform/pkg/events"

// Re-export the subject names so callers in this service import a single
// package rather than reaching into pkg/events directly.
const (
	SubjectUserRegistered = pkgevents.SubjectUserRegistered
	SubjectUserUpdated    = pkgevents.SubjectUserUpdated
)

// UserRegistered is published right after a user successfully signs up. The
// notification service (added later) will subscribe to this to send a welcome
// email; an analytics service might count signups; and so on — none of which
// the user service needs to know about.
type UserRegistered struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserUpdated is published when a profile changes.
type UserUpdated struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
