package edgecontext

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/reddit/baseplate.go/detach"
	"github.com/reddit/baseplate.go/ecinterface"
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

var (
	serializerPool   = thrift.NewTSerializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
	deserializerPool = thrift.NewTDeserializerPoolSizeFactory(1024, thrift.NewTBinaryProtocolFactoryDefault())
)

type v0Impl struct {
	store     *secrets.Store
	keysValue atomic.Value
}

var _ ecinterface.Interface = (*Impl)(nil)

func (impl *v0Impl) ContextToHeader(ctx context.Context) (header string, ok bool) {
	ec, ok := v0GetEdgeContext(ctx)
	if !ok {
		return "", false
	}
	return ec.Header(), true
}

func (impl *v0Impl) HeaderToContext(ctx context.Context, header string) (context.Context, error) {
	ec, err := v0FromHeader(ctx, header, impl)
	if err != nil {
		return ctx, fmt.Errorf("edgecontext.Impl.HeaderToContext: failed to parse header: %w", err)
	}
	return v0SetEdgeContext(ctx, ec), nil
}

type edgeContextKey struct{}

func v0SetEdgeContext(ctx context.Context, ec *v0HeaderUnmarshaler) context.Context {
	if ec == nil {
		return ctx
	}
	return context.WithValue(ctx, edgeContextKey{}, ec)
}

func v0GetEdgeContext(ctx context.Context) (ec *v0HeaderUnmarshaler, ok bool) {
	ec, ok = ctx.Value(edgeContextKey{}).(*v0HeaderUnmarshaler)
	return
}

func v0Factory(cfg Config) ecinterface.Factory {
	return func(args ecinterface.FactoryArgs) (ecinterface.Interface, error) {
		cfg.Store = args.Store
		return Init(cfg), nil
	}
}

func v0Init(cfg Config) *v0Impl {
	impl := &v0Impl{
		store: cfg.Store,
	}
	impl.store.AddMiddlewares(impl.validatorMiddleware)
	ecinterface.Set(impl)
	return impl
}

func v0New(ctx context.Context, impl *v0Impl, args NewArgs) (*v0HeaderUnmarshaler, error) {
	request := ecthrift.NewRequest()
	if args.LoID != "" {
		if !strings.HasPrefix(args.LoID, "t") {
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
	return &v0HeaderUnmarshaler{
		impl:   impl,
		header: header,
		raw:    args,
		ctx:    ctx,
	}, nil
}

func v0FromHeader(ctx context.Context, header string, impl *v0Impl) (*v0HeaderUnmarshaler, error) {
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
	return &v0HeaderUnmarshaler{
		impl:   impl,
		header: header,
		raw:    raw,
		ctx:    ctx,
	}, nil
}

type v0HeaderUnmarshaler struct {
	impl *v0Impl

	// header and raw should always be set during initialization
	header string
	raw    NewArgs

	// token will be validated on first use
	tokenOnce sync.Once
	token     *AuthenticationToken

	// ctx is only used in error logging in AuthToken and UpdateExperimentEvent
	// functions.
	//
	// Since an EdgeContext object is always 1:1 mapped to a request,
	// although storing a ctx object is in general not recommended,
	// it does make sense in this case,
	// and could help us avoiding the awkward situation of needing to pass in ctx
	// object into those functions.
	ctx context.Context
}

var _ HeaderUnmarshaler = (*v0HeaderUnmarshaler)(nil)

func (e *v0HeaderUnmarshaler) getCtx() context.Context {
	if e.ctx != nil {
		return e.ctx
	}
	return context.Background()
}

func (e *v0HeaderUnmarshaler) Unmarshal(data *Data) {
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

func (e *v0HeaderUnmarshaler) AuthenticationToken() *AuthenticationToken {
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

func (e *v0HeaderUnmarshaler) Header() string {
	return e.header
}
