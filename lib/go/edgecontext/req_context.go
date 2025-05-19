package edgecontext

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/reddit/baseplate.go/experiments"
	"github.com/reddit/edgecontext/lib/go/ecdata"
)

// An EdgeRequestContext contains context info about an edge request.
type EdgeRequestContext struct {
	Source ecdata.DataSource

	// ctx is only used in error logging in AuthToken and UpdateExperimentEvent
	// functions.
	//
	// Since an EdgeRequestContext object is always 1:1 mapped to a request,
	// although storing a ctx object is in general not recommended,
	// it does make sense in this case,
	// and could help us avoiding the awkward situation of needing to pass in ctx
	// object into those functions.
	ctx context.Context
}

func NewFromSource(ctx context.Context, source ecdata.DataSource) *EdgeRequestContext {
	if source == nil {
		return nil
	}
	return &EdgeRequestContext{
		Source: source,
		ctx:    ctx,
	}
}

func (e *EdgeRequestContext) getCtx() context.Context {
	if e.ctx != nil {
		return e.ctx
	}
	return context.Background()
}

// AuthToken either validates the raw auth token and cache it,
// or return the cached token.
//
// If the validation failed, the error will be logged.
func (e *EdgeRequestContext) AuthToken() *AuthenticationToken {
	return e.Source.AuthenticationToken()
}

// Header returns the raw, underlying edge request context header that was
// parsed to create the EdgeRequestContext object.
//
// This is not really intended to be used directly but to allow us to propogate
// the header between services.
func (e *EdgeRequestContext) Header() string {
	return e.Source.Header()
}

// SessionID returns the session id of this request.
func (e *EdgeRequestContext) SessionID() string {
	return e.Source.SessionID()
}

// DeviceID returns the device id of this request.
func (e *EdgeRequestContext) DeviceID() string {
	return e.Source.DeviceID()
}

// User returns the info about the user of this request.
func (e *EdgeRequestContext) User() User {
	return User{
		source: e.Source.User(),
	}
}

// CountryCode returns the two-character ISO 3166-1 country code where the
// request orginated from.
func (e *EdgeRequestContext) CountryCode() string {
	return e.Source.CountryCode()
}

// LocaleCode returns the IETF language code for the client
func (e *EdgeRequestContext) LocaleCode() string {
	return e.Source.LocaleCode()
}

// OriginService returns the info about the origin of this request.
func (e *EdgeRequestContext) OriginService() OriginService {
	return OriginService{
		source: e.Source.OriginService(),
	}
}

// OAuthClient returns the info about the oauth client of this request.
//
// ok will be false if this request does not have a valid auth token.
func (e *EdgeRequestContext) OAuthClient() (client OAuthClient, ok bool) {
	token := e.AuthToken()
	if token == nil {
		return
	}
	return OAuthClient(*token), true
}

// Service returns the info about the client service of this request.
//
// ok will be false if this request does not have a valid auth token.
func (e *EdgeRequestContext) Service() (service Service, ok bool) {
	token := e.AuthToken()
	if token == nil {
		return
	}
	return Service(*token), true
}

// UpdateExperimentEvent updates the passed in experiment event with info from
// this edge request context.
//
// It always updates UserID, LoggedIn, CookieCreatedAt, OAuthClientID,
// SessionID, and DeviceID fields,
// and never touches other fields in experiment event.
//
// The caller should create an experiments.ExperimentEvent object,
// with other non-edge-request related fields already filled,
// call this function to update edge-request related fields updated,
// then pass it to an event logger.
func (e *EdgeRequestContext) UpdateExperimentEvent(ee *experiments.ExperimentEvent) {
	e.User().UpdateExperimentEvent(ee)
	if client, ok := e.OAuthClient(); ok {
		client.UpdateExperimentEvent(ee)
	} else {
		ee.OAuthClientID = ""
	}
	ee.SessionID = e.SessionID()
	if deviceID := e.DeviceID(); deviceID != "" {
		var err error
		ee.DeviceID, err = uuid.FromString(deviceID)
		if err != nil {
			ee.DeviceID = uuid.Nil
			slog.ErrorContext(
				e.getCtx(),
				fmt.Sprintf("Failed to parse device id %q into uuid", deviceID),
				"err", err,
			)
		}
	} else {
		ee.DeviceID = uuid.Nil
	}
}

// OriginService holds metadata about the origin of the request.
type OriginService struct {
	source ecdata.OriginServiceSource
}

// Name returns the name of the service that serves as the origin of the request.
func (os OriginService) Name() string {
	return os.source.Name()
}

// RequestID is the id of this request.
func (e *EdgeRequestContext) RequestID() string {
	return e.Source.RequestID()
}
