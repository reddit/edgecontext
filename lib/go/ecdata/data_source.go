package ecdata

import (
	"context"
	"time"
)

type DataSource interface {
	Header() string
	Populate(data *Data)
}

type Data struct {
	AuthenticationToken     func() *AuthenticationToken
	CountryCode             string
	DeviceID                string
	LocaleCode              string
	RequestID               string
	SessionID               string
	OriginServiceName       string
	InsecureLoID            string
	InsecureCookieCreatedAt time.Time
}

type dataSourceContextKey struct{}

func SetDataSource(ctx context.Context, ds DataSource) context.Context {
	if ds == nil {
		return ctx
	}
	return context.WithValue(ctx, dataSourceContextKey{}, ds)
}

func GetDataSource(ctx context.Context) (DataSource, bool) {
	ds, ok := ctx.Value(dataSourceContextKey{}).(DataSource)
	return ds, ok
}
