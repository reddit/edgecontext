package v0

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/reddit/edgecontext/lib/go/ecdata"
)

// An DataSource contains context info about an edge request.
type DataSource struct {
	impl *Impl

	// header and raw should always be set during initialization
	header string
	raw    NewArgs

	// token will be validated on first use
	tokenOnce sync.Once
	token     *ecdata.AuthenticationToken

	// ctx is only used in error logging in AuthToken and UpdateExperimentEvent
	// functions.
	//
	// Since an DataSource object is always 1:1 mapped to a request,
	// although storing a ctx object is in general not recommended,
	// it does make sense in this case,
	// and could help us avoiding the awkward situation of needing to pass in ctx
	// object into those functions.
	ctx context.Context
}

var _ ecdata.DataSource = (*DataSource)(nil)

func (e *DataSource) getCtx() context.Context {
	if e.ctx != nil {
		return e.ctx
	}
	return context.Background()
}

// AuthToken either validates the raw auth token and cache it,
// or return the cached token.
//
// If the validation failed, the error will be logged.
func (e *DataSource) AuthenticationToken() *ecdata.AuthenticationToken {
	e.tokenOnce.Do(func() {
		if token, err := e.impl.ValidateToken(e.raw.AuthToken); err != nil {
			// empty jwt token is considered "normal", no need to spam them in logs.
			if !errors.Is(err, ErrEmptyToken) {
				slog.ErrorContext(
					e.getCtx(), "token validation failed",
					"err", err,
				)
			}
			e.token = nil
		} else {
			e.token = token
		}
	})
	return e.token
}

// Header returns the raw, underlying edge request context header that was
// parsed to create the DataSource object.
//
// This is not really intended to be used directly but to allow us to propogate
// the header between services.
func (e *DataSource) Header() string {
	return e.header
}

// SessionID returns the session id of this request.
func (e *DataSource) SessionID() string {
	return e.raw.SessionID
}

// DeviceID returns the device id of this request.
func (e *DataSource) DeviceID() string {
	return e.raw.DeviceID
}

// User returns the info about the user of this request.
func (e *DataSource) User() ecdata.UserDataSource {
	return User{
		e: e,
	}
}

// CountryCode returns the two-character ISO 3166-1 country code where the
// request orginated from.
func (e *DataSource) CountryCode() string {
	return e.raw.CountryCode
}

// LocaleCode returns the IETF language code for the client
func (e *DataSource) LocaleCode() string {
	return e.raw.LocaleCode
}

// RequestID is the id of this request.
func (e *DataSource) RequestID() string {
	return e.raw.RequestID
}

// OriginService returns the info about the origin of this request.
func (e *DataSource) OriginService() ecdata.OriginServiceSource {
	return OriginService{
		raw: e.raw,
	}
}

// OriginService holds metadata about the origin of the request.
type OriginService struct {
	raw NewArgs
}

// Name returns the name of the service that serves as the origin of the request.
func (os OriginService) Name() string {
	return os.raw.OriginServiceName
}

const userPrefix = "t2_"

type User struct {
	e *DataSource
}

func (u User) AuthenticationToken() *ecdata.AuthenticationToken {
	return u.e.AuthenticationToken()
}

func (u User) InsecureLoID() string {
	return u.e.raw.LoID
}

func (u User) InsecureCookieCreatedAt() time.Time {
	return u.e.raw.LoIDCreatedAt
}
