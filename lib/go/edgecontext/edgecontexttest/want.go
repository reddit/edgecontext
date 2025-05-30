package edgecontexttest

import (
	"context"
	"time"

	"github.com/reddit/baseplate.go/timebp"
	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

type authTokenConfig struct {
	token *edgecontext.AuthenticationToken
}

// AuthenticationTokenOption is a function that configures the WithAuthenticationToken option for Want.
type AuthenticationTokenOption func(*authTokenConfig)

// WithLoggedInUser sets the AuthenticationToken for a logged-in user.
func WithLoggedInUser(id string, createdAt time.Time, roles []string) AuthenticationTokenOption {
	return func(cfg *authTokenConfig) {
		cfg.token.RegisteredClaims.Subject = id
		cfg.token.Roles = roles
		cfg.token.LoID = struct {
			ID        string                      `json:"id,omitempty"`
			CreatedAt timebp.TimestampMillisecond `json:"created_ms,omitempty"`
		}{
			ID:        id,
			CreatedAt: timebp.TimestampMillisecond(createdAt),
		}
	}
}

// WithLoggedOutUser sets the AuthenticationToken for a logged-out user.
func WithLoggedOutUser(loid string, createdAt time.Time, roles []string) AuthenticationTokenOption {
	return func(cfg *authTokenConfig) {
		cfg.token.RegisteredClaims.Subject = ""
		cfg.token.Roles = roles
		cfg.token.LoID = struct {
			ID        string                      `json:"id,omitempty"`
			CreatedAt timebp.TimestampMillisecond `json:"created_ms,omitempty"`
		}{
			ID:        loid,
			CreatedAt: timebp.TimestampMillisecond(createdAt),
		}
	}
}

type serviceConfig struct {
	*authTokenConfig
}

// WithServiceOption is a function that configures the WithService option for WithAuthenticationToken.
type WithServiceOption func(*serviceConfig)

// WithServiceRequestsElevatedAccess sets the AuthenticationToken to indicate that the service requests elevated access.
func WithServiceRequestsElevatedAccess() WithServiceOption {
	return func(cfg *serviceConfig) {
		cfg.token.ServiceRequestedElevatedAccess = true
	}
}

// WithOnBehalfOfUser sets the AuthenticationToken to indicate that the service is acting on behalf of a user.
func WithOnBehalfOfUser(id string, roles []string) WithServiceOption {
	return func(cfg *serviceConfig) {
		cfg.token.OnBehalfOf = &struct {
			AccountID string   `json:"aid,omitempty"`
			Roles     []string `json:"roles,omitempty"`
		}{
			AccountID: id,
			Roles:     roles,
		}
	}
}

// WithService sets the AuthenticationToken for a service.
func WithService(name string, opts ...WithServiceOption) AuthenticationTokenOption {
	return func(cfg *authTokenConfig) {
		cfg.token.RegisteredClaims.Subject = "service/" + name

		serviceCfg := serviceConfig{cfg}
		for _, opt := range opts {
			opt(&serviceCfg)
		}
	}
}

// WithNoAuthenticationToken sets the AuthenticationToken to nil.
func WithNoAuthenticationToken() WantOption {
	return func(cfg *wantConfig) {
		cfg.source.token = nil
	}
}

// WithAuthenticationToken sets the AuthenticationToken in the Want configuration.
func WithAuthenticationToken(opts ...AuthenticationTokenOption) WantOption {
	return func(cfg *wantConfig) {
		authCfg := authTokenConfig{
			token: &edgecontext.AuthenticationToken{},
		}
		for _, opt := range opts {
			opt(&authCfg)
		}
		cfg.source.token = authCfg.token
	}
}

type wantConfig struct {
	source *StubHeaderUnmarshaler
}

// WantOption is a function that configures the Want function.
type WantOption func(*wantConfig)

// WithLocation sets the locale and country codes in the Want configuration.
func WithLocation(localeCode, countryCode string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.localeCode = localeCode
		cfg.source.countryCode = countryCode
	}
}

// WithOriginService sets the origin service name in the Want configuration.
func WithOriginService(name string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.originService = name
	}
}

// WithRequestMetadata sets the request metadata in the Want configuration.
func WithRequestMetadata(requestID, sessionID, deviceID string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.requestID = requestID
		cfg.source.sessionID = sessionID
		cfg.source.deviceID = deviceID
	}
}

// WithLoIDCookie sets the non-authentication token LoID and its creation time in the Want configuration.
func WithLoIDCookie(loid string, createdAt time.Time) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.loid = loid
		cfg.source.cookieCreatedAt = createdAt
	}
}

// WithOAuthClient sets the OAuth client ID and type in the Want configuration.
func WithOAuthClient(id, clientType string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.token.OAuthClientID = id
		cfg.source.token.OAuthClientType = clientType
	}
}

// Want creates a new edgecontext.EdgeRequestContext to compare against in tests.
func Want(opts ...WantOption) *edgecontext.EdgeRequestContext {
	cfg := wantConfig{source: &StubHeaderUnmarshaler{}}
	for _, opt := range opts {
		opt(&cfg)
	}
	return edgecontext.NewFromSource(context.Background(), cfg.source)
}
