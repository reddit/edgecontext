package edgecontext

import (
	"context"
	"time"

	"github.com/reddit/baseplate.go/ecinterface"
	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
	"github.com/reddit/edgecontext/lib/go/ecdata"
)

// LoIDPrefix is the prefix for all LoIDs.
const LoIDPrefix = ecdata.LoIDPrefix

// LocaleRegex validates that locale codes are correctly formatted. They can contain
// either a language, or a language and region specifier separated by an underscore.
// e.g. en, en_US
var LocaleRegex = ecdata.LocaleRegex

var (
	// ErrLoIDWrongPrefix is an error could be returned by New() when passed in LoID
	// does not have the correct prefix.
	ErrLoIDWrongPrefix = ecdata.ErrLoIDWrongPrefix

	// ErrInvalidLocaleCode is returned by New() when an invalid locale code is passed in.
	ErrInvalidLocaleCode = ecdata.ErrInvalidLocaleCode
)

// An Impl is an initialized edge context implementation.
//
// It implements ecinterface.Interface.
//
// Please call Init function to initialize it.
type Impl struct {
	ecinterface.Interface
}

type dataSourceContextKey struct{}

// SetEdgeContext sets the given EdgeRequestContext on the context object.
func SetEdgeContext(ctx context.Context, ec *EdgeRequestContext) context.Context {
	return ecdata.SetDataSource(ctx, ec.Source())
}

// GetEdgeContext gets the current EdgeRequestContext from the context object,
// if set.
func GetEdgeContext(ctx context.Context) (*EdgeRequestContext, bool) {
	source, ok := ecdata.GetDataSource(ctx)
	if !ok {
		return nil, false
	}

	return NewFromSource(ctx, source), true
}

// Config for Init function.
type Config struct {
	// The secret store to get the keys for jwt validation
	Store *secrets.Store

	// The logger to log key decoding errors
	//
	// deprecated: this is not used internally, slog should be used instead
	Logger log.Wrapper
}

// Factory returns an ecinterface.Factory implementation by wrapping Init.
//
// The Store in cfg will be replaced by the Factory arg.
func Factory(cfg Config) ecinterface.Factory {
	return getBackend().Factory(cfg)
}

// Init intializes an Impl.
//
// It also calls ecinterface.Set to store the implementation created globally.
func Init(cfg Config) *Impl {
	return getBackend().Init(cfg)
}

// NewArgs are the args for New function.
//
// All fields are optional.
type NewArgs struct {
	// If LoID is non-empty, it must have prefix of LoIDPrefix ("t2_").
	LoID          string
	LoIDCreatedAt time.Time

	SessionID string

	DeviceID string

	AuthToken string

	OriginServiceName string

	CountryCode string

	RequestID string

	LocaleCode string
}

// New creates a new EdgeRequestContext from scratch.
//
// This function should be used by services on the edge talking to clients
// directly, after talked to authentication service to get the auth token.
func New(ctx context.Context, impl *Impl, args NewArgs) (*EdgeRequestContext, error) {
	return getBackend().New(ctx, impl, args)
}

// FromHeader returns a new EdgeRequestContext from the given header string
// using the given Impl.
func FromHeader(ctx context.Context, header string, impl *Impl) (*EdgeRequestContext, error) {
	return getBackend().FromHeader(ctx, header, impl)
}
