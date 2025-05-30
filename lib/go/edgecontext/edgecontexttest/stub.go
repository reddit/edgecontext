package edgecontexttest

import (
	"time"

	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

// StubHeaderUnmarshaler is a stub implementation of edgecontext.HeaderUnmarshaler that can be used in tests.
type StubHeaderUnmarshaler struct {
	token *edgecontext.AuthenticationToken

	countryCode   string
	localeCode    string
	deviceID      string
	requestID     string
	sessionID     string
	originService string

	loid            string
	cookieCreatedAt time.Time
}

// Unmarshal function sets the fields of the provided edgecontext.Data.
func (s *StubHeaderUnmarshaler) Unmarshal(data *edgecontext.Data) {
	data.AuthenticationToken = func() *edgecontext.AuthenticationToken {
		return s.token
	}
	data.CountryCode = s.countryCode
	data.DeviceID = s.deviceID
	data.LocaleCode = s.localeCode
	data.RequestID = s.requestID
	data.SessionID = s.sessionID
	data.OriginServiceName = s.originService
	data.InsecureLoID = s.loid
	data.InsecureCookieCreatedAt = s.cookieCreatedAt
}

// Header is always returns an empty string in this stub implementation.
func (s *StubHeaderUnmarshaler) Header() string {
	return ""
}
