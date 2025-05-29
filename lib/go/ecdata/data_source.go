package ecdata

import (
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
