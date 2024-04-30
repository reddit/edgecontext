package edgecontext

import (
	"strings"
)

const servicePrefix = "service/"

// A Service wraps AuthenticationToken and provides info about an authenticated
// service talking to us.
type Service AuthenticationToken

func (s Service) isService() bool {
	subject := AuthenticationToken(s).Subject()
	return strings.HasPrefix(subject, servicePrefix)
}

// Name returns the name of the service.
//
// If it's not coming from an authenticated service,
// ("", false) will be returned.
func (s Service) Name() (name string, ok bool) {
	if s.isService() {
		subject := AuthenticationToken(s).Subject()
		return subject[len(servicePrefix):], true
	}
	return
}

func (s Service) OnBehalfOfID() (id string, ok bool) {
	if s.isService() {
		token := AuthenticationToken(s)
		if token.OnBehalfOf == nil {
			return
		}
		if strings.HasPrefix(token.OnBehalfOf.AccountID, userPrefix) {
			return token.OnBehalfOf.AccountID, true
		}
		return
	}
	return
}

func (s Service) OnBehalfOfRoles() (roles []string, ok bool) {
	if s.isService() {
		token := AuthenticationToken(s)
		if token.OnBehalfOf == nil {
			return
		}
		return token.OnBehalfOf.Roles, true
	}
	return
}

func (s Service) IsElevatedAccess() bool {
	if s.isService() {
		return AuthenticationToken(s).ServiceRequestedElevatedAccess
	}
	return false
}
