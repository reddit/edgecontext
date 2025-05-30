package edgecontexttest

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

type user struct {
	ID         string
	IsLoggedIn bool
	LoID       string
	CreatedAt  time.Time
	Roles      []string
}

type originService struct {
	Name string
}

type oauthClient struct {
	ID   string
	Type string
}

type service struct {
	Name                   string
	RequestsElevatedAccess bool
	OnBehalfOf             user
}

type edgeContext struct {
	HasAuthenticationToken bool
	RequestID              string
	SessionID              string
	DeviceID               string
	LocaleCode             string
	CountryCode            string

	User          user
	OriginService originService
	OAuthClient   oauthClient
	Service       service
}

// Diff compares two EdgeRequestContext objects and returns a string if any differences are found.
func Diff(want, got *edgecontext.EdgeRequestContext) string {
	return cmp.Diff(
		want, got,
		transform(),
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	)
}

func transform() cmp.Option {
	return cmp.Transformer("EdgeRequestContext", func(ec *edgecontext.EdgeRequestContext) edgeContext {
		transformed := edgeContext{
			HasAuthenticationToken: ec.AuthToken() != nil,
			RequestID:              ec.RequestID(),
			SessionID:              ec.SessionID(),
			DeviceID:               ec.DeviceID(),
			LocaleCode:             ec.LocaleCode(),
			CountryCode:            ec.CountryCode(),

			OriginService: originService{
				Name: ec.OriginService().Name(),
			},
		}

		if id, ok := ec.User().ID(); ok {
			loid, _ := ec.User().LoID()
			createdAt, _ := ec.User().CookieCreatedAt()
			transformed.User = user{
				ID:         id,
				IsLoggedIn: ec.User().IsLoggedIn(),
				LoID:       loid,
				CreatedAt:  createdAt,
				Roles:      ec.User().Roles(),
			}
		} else {
			loid, _ := ec.User().LoID()
			createdAt, _ := ec.User().CookieCreatedAt()
			transformed.User = user{
				ID:         "",
				IsLoggedIn: ec.User().IsLoggedIn(),
				LoID:       loid,
				CreatedAt:  createdAt,
				Roles:      ec.User().Roles(),
			}
		}

		if client, ok := ec.OAuthClient(); ok {
			transformed.OAuthClient = oauthClient{
				ID:   client.ID(),
				Type: ec.AuthToken().OAuthClientType,
			}
		}

		if svc, ok := ec.Service(); ok {
			name, _ := svc.Name()
			transformed.Service = service{
				Name:                   name,
				RequestsElevatedAccess: svc.RequestsElevatedAccess(),
			}
			if onBehalfOf, ok := svc.OnBehalfOfID(); ok {
				roles, _ := svc.OnBehalfOfRoles()
				transformed.Service.OnBehalfOf = user{
					ID:    onBehalfOf,
					Roles: roles,
				}
			}
		}
		return transformed
	})
}
