package edgecontexttest

import (
	"context"
	"time"

	"github.com/reddit/baseplate.go/timebp"
	"github.com/reddit/edgecontext/lib/go/ecdata"
	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

type authTokenConfig struct {
	token *ecdata.AuthenticationToken
}

type AuthenticationTokenOption func(*authTokenConfig)

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

type WithServiceOption func(*serviceConfig)

func WithServiceRequestsElevatedAccess() WithServiceOption {
	return func(cfg *serviceConfig) {
		cfg.token.ServiceRequestedElevatedAccess = true
	}
}

func WithOnBehalfOfUser(id string, roles []string) WithServiceOption {
	return func(cfg *serviceConfig) {
		cfg.token.OnBehalfOf = &ecdata.OnBehalfOf{
			AccountID: id,
			Roles:     roles,
		}
	}
}

func WithService(name string, opts ...WithServiceOption) AuthenticationTokenOption {
	return func(cfg *authTokenConfig) {
		cfg.token.RegisteredClaims.Subject = "service/" + name

		serviceCfg := serviceConfig{cfg}
		for _, opt := range opts {
			opt(&serviceCfg)
		}
	}
}

func WithNoAuthenticationToken() WantOption {
	return func(cfg *wantConfig) {
		cfg.source.token = nil
	}
}

func WithAuthenticationToken(opts ...AuthenticationTokenOption) WantOption {
	return func(cfg *wantConfig) {
		authCfg := authTokenConfig{
			token: &ecdata.AuthenticationToken{},
		}
		for _, opt := range opts {
			opt(&authCfg)
		}
		cfg.source.token = authCfg.token
	}
}

type wantConfig struct {
	source *StubDataSource
}

type WantOption func(*wantConfig)

func WithLocation(localeCode, countryCode string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.localeCode = localeCode
		cfg.source.countryCode = countryCode
	}
}

func WithOriginService(name string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.originService = name
	}
}

func WithRequestMetadata(requestID, sessionID, deviceID string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.requestID = requestID
		cfg.source.sessionID = sessionID
		cfg.source.deviceID = deviceID
	}
}

func WithLoIDCookie(loid string, createdAt time.Time) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.loid = loid
		cfg.source.cookieCreatedAt = createdAt
	}
}

func WithOAuthClient(id, clientType string) WantOption {
	return func(cfg *wantConfig) {
		cfg.source.token.OAuthClientID = id
		cfg.source.token.OAuthClientType = clientType
	}
}

func Want(opts ...WantOption) *edgecontext.EdgeRequestContext {
	cfg := wantConfig{source: &StubDataSource{}}
	for _, opt := range opts {
		opt(&cfg)
	}
	return edgecontext.NewFromSource(context.Background(), cfg.source)
}
