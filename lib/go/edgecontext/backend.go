package edgecontext

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/reddit/baseplate.go/ecinterface"
	v0 "github.com/reddit/edgecontext/lib/go/internal/v0"
)

func init() {
	SetBackend(defaultBackend())
}

var backend atomic.Pointer[Backend]

type Backend struct {
	SetEdgeContext func(context.Context, *EdgeRequestContext) context.Context
	GetEdgeContext func(context.Context) (*EdgeRequestContext, bool)
	Factory        func(Config) ecinterface.Factory
	Init           func(Config) *Impl
	New            func(context.Context, *Impl, NewArgs) (*EdgeRequestContext, error)
	FromHeader     func(context.Context, string, *Impl) (*EdgeRequestContext, error)
}

func defaultBackend() *Backend {
	return &Backend{
		SetEdgeContext: func(ctx context.Context, ec *EdgeRequestContext) context.Context {
			if ec == nil {
				return ctx
			}
			return v0.SetEdgeContext(ctx, ec.Source.(*v0.DataSource))
		},
		GetEdgeContext: func(ctx context.Context) (*EdgeRequestContext, bool) {
			source, ok := v0.GetEdgeContext(ctx)
			if !ok {
				return nil, false
			}
			return &EdgeRequestContext{
				Source: source,
				ctx:    ctx,
			}, true
		},
		Factory: func(cfg Config) ecinterface.Factory {
			return v0.Factory(v0.Config{
				Store: cfg.Store,
			})
		},
		Init: func(cfg Config) *Impl {
			return &Impl{
				Interface: v0.Init(v0.Config{
					Store: cfg.Store,
				}),
			}
		},
		New: func(ctx context.Context, impl *Impl, args NewArgs) (*EdgeRequestContext, error) {
			unwrapped, ok := impl.Interface.(*v0.Impl)
			if !ok {
				return nil, fmt.Errorf("unwrapped interface is not a v0.Impl")
			}
			source, err := v0.New(
				ctx,
				unwrapped,
				v0.NewArgs{
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
			unwrapped, ok := impl.Interface.(*v0.Impl)
			if !ok {
				return nil, fmt.Errorf("unwrapped interface is not a v0.Impl")
			}
			source, err := v0.FromHeader(ctx, header, unwrapped)
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

func SetBackend(b *Backend) {
	if b == nil {
		panic("backend cannot be nil")
	}
	if b.SetEdgeContext == nil {
		panic("SetEdgeContext cannot be nil")
	}
	if b.GetEdgeContext == nil {
		panic("GetEdgeContext cannot be nil")
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
