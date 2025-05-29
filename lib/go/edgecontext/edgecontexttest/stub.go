package edgecontexttest

import (
	"time"

	"github.com/reddit/edgecontext/lib/go/ecdata"
)

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
}

func (s *StubDataSource) Populate(data *ecdata.Data) {
	data.AuthenticationToken = func() *ecdata.AuthenticationToken {
		return s.token
	}
	data.CountryCode = s.countryCode
	data.DeviceID = s.deviceID
	data.LocaleCode = s.localeCode
	data.RequestID = s.requestID
	data.SessionID = s.sessionID
	data.OriginServiceName = s.originService
	data.InsecureLoID = s.loid
	data.InsecureCookieCreatedAt = s.cookieCreatedAt
}

func (s *StubDataSource) Header() string {
	return ""
}
