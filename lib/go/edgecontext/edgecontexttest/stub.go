package edgecontexttest

import (
	"time"

	"github.com/reddit/edgecontext/lib/go/ecdata"
)

type Logger interface {
	Logf(format string, args ...interface{})
}

type StubDataSource struct {
	token *ecdata.AuthenticationToken

	countryCode   string
	localeCode    string
	deviceID      string
	requestID     string
	sessionID     string
	originService string

	loid            string
	cookieCreatedAt time.Time

	logger Logger
}

func (s *StubDataSource) AuthenticationToken() *ecdata.AuthenticationToken {
	return s.token
}

func (s *StubDataSource) Header() string {
	return ""
}

func (s *StubDataSource) CountryCode() string {
	return s.countryCode
}

func (s *StubDataSource) DeviceID() string {
	return s.deviceID
}

func (s *StubDataSource) LocaleCode() string {
	return s.localeCode
}

func (s *StubDataSource) RequestID() string {
	return s.requestID
}

func (s *StubDataSource) SessionID() string {
	return s.sessionID
}

func (s *StubDataSource) User() ecdata.UserDataSource {
	return &StubUserDataSource{
		token:           s.token,
		loID:            s.loid,
		cookieCreatedAt: s.cookieCreatedAt,
	}
}

func (s *StubDataSource) OriginService() ecdata.OriginServiceSource {
	return &StubOriginServiceDataSource{
		name: s.originService,
	}
}

type StubUserDataSource struct {
	token *ecdata.AuthenticationToken

	loID            string
	cookieCreatedAt time.Time
}

func (s *StubUserDataSource) AuthenticationToken() *ecdata.AuthenticationToken {
	return s.token
}

func (s *StubUserDataSource) InsecureCookieCreatedAt() time.Time {
	return s.cookieCreatedAt
}

func (s *StubUserDataSource) InsecureLoID() string {
	return s.loID
}

type StubOriginServiceDataSource struct {
	name string
}

func (s *StubOriginServiceDataSource) Name() string {
	return s.name
}
