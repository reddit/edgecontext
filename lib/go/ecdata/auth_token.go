package ecdata

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/reddit/baseplate.go/timebp"
)

type LoID struct {
	ID        string                      `json:"id,omitempty"`
	CreatedAt timebp.TimestampMillisecond `json:"created_ms,omitempty"`
}

// OnBehalfOf defines the structure for the "on behalf of" field in the authentication token.
type OnBehalfOf struct {
	AccountID string   `json:"aid,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// AuthenticationToken defines the json format of the authentication token.
type AuthenticationToken struct {
	jwt.RegisteredClaims

	Roles []string `json:"roles,omitempty"`

	OAuthClientID   string   `json:"client_id,omitempty"`
	OAuthClientType string   `json:"client_type,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`

	LoID LoID `json:"loid,omitempty"`
	
	OnBehalfOf                     *OnBehalfOf `json:"obo,omitempty"`
	ServiceRequestedElevatedAccess bool        `json:"sea,omitempty"`
}

// Subject returns the subject field of the token.
func (t AuthenticationToken) Subject() string {
	return t.RegisteredClaims.Subject
}
