package edgecontext

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/reddit/baseplate.go/ecinterface"
)

func init() {
	SetBackend(defaultBackend())
}

var backend atomic.Pointer[Backend]

// Backend provides the interface for the edgecontext package to interact with the underlying implementation.
type Backend struct {
	// Set is the implementation for SetEdgeContext.
	Set func(context.Context, *EdgeRequestContext) context.Context

	// Get is the implementation for GetEdgeContext.
	Get func(context.Context) (*EdgeRequestContext, bool)

	// Factory is the implementation for the Factory function.
	Factory func(Config) ecinterface.Factory

	// Init is the implementation for the Init function.
	Init func(Config) *Impl

	// New is the implementation for New.
	New func(context.Context, *Impl, NewArgs) (*EdgeRequestContext, error)

	// FromHeader is the implementation for FromHeader.
	FromHeader func(context.Context, string, *Impl) (*EdgeRequestContext, error)
}

func defaultBackend() *Backend {
	return &Backend{
		Set: func(ctx context.Context, ec *EdgeRequestContext) context.Context {
			if ec == nil {
				return ctx
			}
			return v0SetEdgeContext(ctx, ec.Source().(*v0HeaderUnmarshaler))
		},
		Get: func(ctx context.Context) (*EdgeRequestContext, bool) {
			source, ok := v0GetEdgeContext(ctx)
			if !ok {
				return nil, false
			}
			return NewFromSource(ctx, source), true
		},
		Factory: func(cfg Config) ecinterface.Factory {
			return v0Factory(Config{
				Store: cfg.Store,
			})
		},
		Init: func(cfg Config) *Impl {
			return &Impl{
				Interface: v0Init(Config{
					Store: cfg.Store,
				}),
			}
		},
		New: func(ctx context.Context, impl *Impl, args NewArgs) (*EdgeRequestContext, error) {
			unwrapped, ok := impl.Interface.(*v0Impl)
			if !ok {
				return nil, fmt.Errorf("unwrapped interface is not a v0Impl")
			}
			source, err := v0New(
				ctx,
				unwrapped,
				NewArgs{
					LoID:              args.LoID,
					LoIDCreatedAt:     args.LoIDCreatedAt,
					SessionID:         args.SessionID,
					DeviceID:          args.DeviceID,
					AuthToken:         args.AuthToken,
					OriginServiceName: args.OriginServiceName,
					CountryCode:       args.CountryCode,
					RequestID:         args.RequestID,
					LocaleCode:        args.LocaleCode,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("creating v0 data Source: %w", err)
			}
			return NewFromSource(ctx, source), nil
		},
		FromHeader: func(ctx context.Context, header string, impl *Impl) (*EdgeRequestContext, error) {
			unwrapped, ok := impl.Interface.(*v0Impl)
			if !ok {
				return nil, fmt.Errorf("unwrapped interface is not a v0Impl")
			}
			source, err := v0FromHeader(ctx, header, unwrapped)
			if err != nil {
				return nil, fmt.Errorf("creating v0 data Source: %w", err)
			}
			if source == nil {
				return nil, nil
			}
			return NewFromSource(ctx, source), nil
		},
	}
}

// SetBackend sets the backend for the edgecontext package. This should only be called once during setup.
func SetBackend(b *Backend) {
	if b == nil {
		panic("backend cannot be nil")
	}
	if b.Set == nil {
		panic("Set cannot be nil")
	}
	if b.Get == nil {
		panic("Get cannot be nil")
	}
	if b.Factory == nil {
		panic("Factory cannot be nil")
	}
	if b.Init == nil {
		panic("Init cannot be nil")
	}
	if b.New == nil {
		panic("New cannot be nil")
	}
	if b.FromHeader == nil {
		panic("FromHeader cannot be nil")
	}
	backend.Store(b)
}

func getBackend() *Backend {
	return backend.Load()
}
