package edgecontext

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/reddit/baseplate.go/ecinterface"
	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
)

// HeaderUnmarshaler is responsible for unmarshaling the edge context header into a Data struct.
type HeaderUnmarshaler interface {
	// Header returns the raw header string that can be unmarshaled.
	Header() string

	// Unmarshal unmarshals the header into the provided Data struct. An EdgeRequestContext will only call Unmarshal once and
	// cache the result.
	Unmarshal(data *Data)
}

// Data contains the unmarshaled data from the edge context header.
type Data struct {
	AuthenticationToken func() *AuthenticationToken
	CountryCode         string
	DeviceID            string
	LocaleCode          string
	RequestID           string
	SessionID           string
	OriginServiceName   string

	// InsecureLoID is the LoID from the loid cookie/header rather than the one from the auth token.
	InsecureLoID string
	// InsecureCookieCreatedAt is the created time of the LoID cookie/header rather than the one from the auth token.
	InsecureCookieCreatedAt time.Time
}

// LoIDPrefix is the prefix for all LoIDs.
const LoIDPrefix = "t2_"

// LocaleRegex validates that locale codes are correctly formatted. They can contain
// either a language, or a language and region specifier separated by an underscore.
// e.g. en, en_US
var LocaleRegex = regexp.MustCompile(`^[a-z]{2,}([_|\-][\da-zA-Z]{2,})*$`)

var (
	// ErrLoIDWrongPrefix is an error could be returned by New() when passed in LoID
	// does not have the correct prefix.
	ErrLoIDWrongPrefix = errors.New("edgecontext: loid should have " + LoIDPrefix + " prefix")

	// ErrInvalidLocaleCode is returned by New() when an invalid locale code is passed in.
	ErrInvalidLocaleCode = errors.New("edgecontext: locale code should match format: en, en_US")
)

// An Impl is an initialized edge context implementation.
//
// It implements ecinterface.Interface.
//
// Please call Init function to initialize it.
type Impl struct {
	ecinterface.Interface
}

// SetEdgeContext sets the given EdgeRequestContext on the context object.
func SetEdgeContext(ctx context.Context, ec *EdgeRequestContext) context.Context {
	return getBackend().Set(ctx, ec)
}

// GetEdgeContext gets the current EdgeRequestContext from the context object,
// if set.
func GetEdgeContext(ctx context.Context) (*EdgeRequestContext, bool) {
	return getBackend().Get(ctx)
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
	return &Impl{
		Interface: getBackend().Init(cfg),
	}
}

func (i *Impl) ValidateToken(token string) (*AuthenticationToken, error) {
	ec, err := New(context.Background(), i, NewArgs{
		AuthToken: token,
	})
	if err != nil {
		return nil, err
	}
	return ec.AuthToken(), nil
}

func (i *Impl) FromHeader(ctx context.Context, header string) (*EdgeRequestContext, error) {
	return getBackend().FromHeader(ctx, header, i)
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
