package edgecontext

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/reddit/baseplate.go/detach"
	"github.com/reddit/baseplate.go/ecinterface"
	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
	"github.com/reddit/baseplate.go/timebp"

	ecthrift "github.com/reddit/edgecontext/lib/go/internal/reddit/edgecontext"
)

func init() {
	copyEC := func(dst, src context.Context) context.Context {
		if ec, ok := GetEdgeContext(src); ok {
			dst = SetEdgeContext(dst, ec)
		}
		return dst
	}

	detach.Register(detach.Hooks{
		Inline: copyEC,
		Async: func(dst, src context.Context, next func(ctx context.Context)) {
			next(copyEC(dst, src))
		},
	})
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
	store     *secrets.Store
	logger    log.Wrapper
	keysValue atomic.Value
}

var _ ecinterface.Interface = (*Impl)(nil)

// ContextToHeader implements ecinterface.Interface.
func (impl *Impl) ContextToHeader(ctx context.Context) (header string, ok bool) {
	ec, ok := GetEdgeContext(ctx)
	if !ok {
		return "", false
	}
	return ec.Header(), true
}

// HeaderToContext implements ecinterface.Interface.
func (impl *Impl) HeaderToContext(ctx context.Context, header string) (context.Context, error) {
	ec, err := FromHeader(ctx, header, impl)
	if err != nil {
		return ctx, fmt.Errorf("edgecontext.Impl.HeaderToContext: failed to parse header: %w", err)
	}
	return SetEdgeContext(ctx, ec), nil
}

var (
	serializerPool   = thrift.NewTSerializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
	deserializerPool = thrift.NewTDeserializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
)

type contextKey int

const (
	edgeContextKey contextKey = iota
)

// SetEdgeContext sets the given EdgeRequestContext on the context object.
func SetEdgeContext(ctx context.Context, ec *EdgeRequestContext) context.Context {
	if ec == nil {
		return ctx
	}
	return context.WithValue(ctx, edgeContextKey, ec)
}

// GetEdgeContext gets the current EdgeRequestContext from the context object,
// if set.
func GetEdgeContext(ctx context.Context) (ec *EdgeRequestContext, ok bool) {
	ec, ok = ctx.Value(edgeContextKey).(*EdgeRequestContext)
	return
}

// Config for Init function.
type Config struct {
	// The secret store to get the keys for jwt validation
	Store *secrets.Store
	// The logger to log key decoding errors
	Logger log.Wrapper
}

// Factory returns an ecinterface.Factory implementation by wrapping Init.
//
// The Store in cfg will be replaced by the Factory arg.
func Factory(cfg Config) ecinterface.Factory {
	return func(args ecinterface.FactoryArgs) (ecinterface.Interface, error) {
		cfg.Store = args.Store
		return Init(cfg), nil
	}
}

// Init intializes an Impl.
//
// It also calls ecinterface.Set to store the implementation created globally.
func Init(cfg Config) *Impl {
	impl := &Impl{
		store:  cfg.Store,
		logger: cfg.Logger,
	}
	impl.store.AddMiddlewares(impl.validatorMiddleware)
	ecinterface.Set(impl)
	return impl
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
	request := ecthrift.NewRequest()
	if args.LoID != "" {
		if !strings.HasPrefix(args.LoID, userPrefix) {
			return nil, ErrLoIDWrongPrefix
		}
		request.Loid = &ecthrift.Loid{
			ID:        args.LoID,
			CreatedMs: timebp.TimeToMilliseconds(args.LoIDCreatedAt),
		}
	}
	if args.SessionID != "" {
		request.Session = &ecthrift.Session{
			ID: args.SessionID,
		}
	}
	if args.DeviceID != "" {
		request.Device = &ecthrift.Device{
			ID: args.DeviceID,
		}
	}
	if args.OriginServiceName != "" {
		request.OriginService = &ecthrift.OriginService{
			Name: args.OriginServiceName,
		}
	}
	if args.CountryCode != "" {
		request.Geolocation = &ecthrift.Geolocation{
			CountryCode: ecthrift.CountryCode(args.CountryCode),
		}
	}
	if args.RequestID != "" {
		request.RequestID = &ecthrift.RequestId{
			ReadableID: args.RequestID,
		}
	}
	if args.LocaleCode != "" {
		if !LocaleRegex.MatchString(args.LocaleCode) {
			return nil, ErrInvalidLocaleCode
		}
		request.Locale = &ecthrift.Locale{
			LocaleCode: ecthrift.LocaleCode(args.LocaleCode),
		}
	}

	request.AuthenticationToken = ecthrift.AuthenticationToken(args.AuthToken)

	header, err := serializerPool.WriteString(ctx, request)
	if err != nil {
		return nil, err
	}
	return &EdgeRequestContext{
		impl:   impl,
		header: header,
		raw:    args,
		ctx:    ctx,
	}, nil
}

// FromHeader returns a new EdgeRequestContext from the given header string
// using the given Impl.
func FromHeader(ctx context.Context, header string, impl *Impl) (*EdgeRequestContext, error) {
	if header == "" {
		return nil, nil
	}

	request := ecthrift.NewRequest()
	if err := deserializerPool.ReadString(ctx, request, header); err != nil {
		return nil, err
	}

	raw := NewArgs{
		AuthToken: string(request.AuthenticationToken),
	}
	if request.Session != nil {
		raw.SessionID = request.Session.ID
	}
	if request.Device != nil {
		raw.DeviceID = request.Device.ID
	}
	if request.Loid != nil {
		raw.LoID = request.Loid.ID
		raw.LoIDCreatedAt = timebp.MillisecondsToTime(request.Loid.CreatedMs)
	}
	if request.OriginService != nil {
		raw.OriginServiceName = request.OriginService.Name
	}
	if request.Geolocation != nil {
		raw.CountryCode = string(request.Geolocation.CountryCode)
	}
	if request.RequestID != nil {
		raw.RequestID = request.RequestID.ReadableID
	}
	if request.Locale != nil {
		raw.LocaleCode = string(request.Locale.LocaleCode)
	}
	return &EdgeRequestContext{
		impl:   impl,
		header: header,
		raw:    raw,
		ctx:    ctx,
	}, nil
}
