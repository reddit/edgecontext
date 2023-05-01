package edgecontext

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/reddit/baseplate.go/timebp"
)

// AuthenticationToken defines the json format of the authentication token.
type AuthenticationToken struct {
	jwt.RegisteredClaims

	// NOTE: Subject field is in StandardClaims.

	Roles []string `json:"roles,omitempty"`

	OAuthClientID   string   `json:"client_id,omitempty"`
	OAuthClientType string   `json:"client_type,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`

	LoID struct {
		ID        string                      `json:"id,omitempty"`
		CreatedAt timebp.TimestampMillisecond `json:"created_ms,omitempty"`
	} `json:"loid,omitempty"`
}

// Subject returns the subject field of the token.
func (t AuthenticationToken) Subject() string {
	return t.RegisteredClaims.Subject
}
