package ecdata

import (
	"time"
)

type AuthenticationTokenSource interface {
	AuthenticationToken() *AuthenticationToken
}
type UserDataSource interface {
	AuthenticationTokenSource

	InsecureCookieCreatedAt() time.Time
	InsecureLoID() string
}

type OriginServiceSource interface {
	Name() string
}
type DataSource interface {
	AuthenticationTokenSource

	Header() string

	CountryCode() string
	DeviceID() string
	LocaleCode() string
	RequestID() string
	SessionID() string

	User() UserDataSource
	OriginService() OriginServiceSource
}
