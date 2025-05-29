package v0

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/reddit/baseplate.go/detach"
	"github.com/reddit/baseplate.go/ecinterface"
	"github.com/reddit/baseplate.go/secrets"
	"github.com/reddit/baseplate.go/timebp"
	"github.com/reddit/edgecontext/lib/go/ecdata"

	ecthrift "github.com/reddit/edgecontext/lib/go/internal/reddit/edgecontext"
)

func init() {
	copyEC := func(dst, src context.Context) context.Context {
		if ec, ok := ecdata.GetDataSource(src); ok {
			dst = ecdata.SetDataSource(dst, ec)
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

const LoIDPrefix = ecdata.LoIDPrefix

var LocaleRegex = regexp.MustCompile(`^[a-z]{2,}([_|\-][\da-zA-Z]{2,})*$`)

var (
	serializerPool   = thrift.NewTSerializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
	deserializerPool = thrift.NewTDeserializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
)

// An Impl is an initialized edge context implementation.
//
// It implements ecinterface.Interface.
//
// Please call Init function to initialize it.
type Impl struct {
	store     *secrets.Store
	keysValue atomic.Value
}

var _ ecinterface.Interface = (*Impl)(nil)

// ContextToHeader implements ecinterface.Interface.
func (impl *Impl) ContextToHeader(ctx context.Context) (header string, ok bool) {
	source, ok := ecdata.GetDataSource(ctx)
	if !ok {
		return "", false
	}
	return source.Header(), true
}

// HeaderToContext implements ecinterface.Interface.
func (impl *Impl) HeaderToContext(ctx context.Context, header string) (context.Context, error) {
	ec, err := FromHeader(ctx, header, impl)
	if err != nil {
		return ctx, fmt.Errorf("edgecontext.Impl.HeaderToContext: failed to parse header: %w", err)
	}
	return ecdata.SetDataSource(ctx, ec), nil
}

// Config for Init function.
type Config struct {
	// The secret store to get the keys for jwt validation
	Store *secrets.Store
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
		store: cfg.Store,
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

// New creates a new DataSource from scratch.
//
// This function should be used by services on the edge talking to clients
// directly, after talked to authentication service to get the auth token.
func New(ctx context.Context, impl *Impl, args NewArgs) (*DataSource, error) {
	request := ecthrift.NewRequest()
	if args.LoID != "" {
		if !strings.HasPrefix(args.LoID, "t2_") {
			return nil, ecdata.ErrLoIDWrongPrefix
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
			return nil, ecdata.ErrInvalidLocaleCode
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
	return &DataSource{
		impl:   impl,
		header: header,
		raw:    args,
		ctx:    ctx,
	}, nil
}

// FromHeader returns a new DataSource from the given header string
// using the given Impl.
func FromHeader(ctx context.Context, header string, impl *Impl) (*DataSource, error) {
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
	return &DataSource{
		impl:   impl,
		header: header,
		raw:    raw,
		ctx:    ctx,
	}, nil
}
