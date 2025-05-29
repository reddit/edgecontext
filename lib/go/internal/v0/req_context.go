package v0

import (
	"context"
	"errors"
	"log/slog"
	"sync"

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

// Populate populates the given ecdata.Data object with the data from this DataSource.
func (e *DataSource) Populate(data *ecdata.Data) {
	data.AuthenticationToken = e.AuthenticationToken
	data.SessionID = e.raw.SessionID
	data.DeviceID = e.raw.DeviceID
	data.CountryCode = e.raw.CountryCode
	data.LocaleCode = e.raw.LocaleCode
	data.RequestID = e.raw.RequestID
	data.OriginServiceName = e.raw.OriginServiceName
	data.InsecureLoID = e.raw.LoID
	data.InsecureCookieCreatedAt = e.raw.LoIDCreatedAt
}

// AuthenticationToken either validates the raw auth token and cache it,
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
// parsed to create the edge context.
func (e *DataSource) Header() string {
	return e.header
}
