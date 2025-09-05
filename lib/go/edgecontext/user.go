package edgecontext

import (
	"slices"
	"strings"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/reddit/baseplate.go/experiments"
	"github.com/reddit/baseplate.go/timebp"
)

const userPrefix = "t2_"

// AccountType corresponds to the account_type field of the database account
// It can have be present for both logged-in and logged-out users.
const (
	AccountTypeUnknown        = "UNKNOWN"
	AccountTypeUser           = "" // historically the empty string has meant a regular user
	AccountTypeProxy          = "PROXY"
	AccountTypeApp            = "APP"
	AccountTypeBrand          = "BRAND"
	AccountTypeAdsIntegration = "ADS_INTEGRATION"
	AccountTypeLite           = "LITE"
)

// An User wraps *EdgeRequestContext and provides info about a logged in or
// logged our user.
type User struct {
	data func() Data
}

// ID returns the authenticated account id of the user.
//
// ok will be false if the user is not logged in.
func (u User) ID() (id string, ok bool) {
	token := u.data().AuthenticationToken()
	if token == nil {
		return
	}
	if strings.HasPrefix(token.Subject(), userPrefix) {
		return token.Subject(), true
	}
	return
}

// IsLoggedIn returns true if the user is logged in.
func (u User) IsLoggedIn() bool {
	_, ok := u.ID()
	return ok
}

// LoID returns the LoID of this user.
func (u User) LoID() (loid string, ok bool) {
	// First, we return the logged in user id if it's a logged in user.
	if id, ok := u.ID(); ok {
		return id, ok
	}

	// Then, we use the loid from the thrift payload.
	if insecure := u.data().InsecureLoID; insecure != "" {
		return insecure, true
	}

	// Finally, we fallback to the loid from the JWT token.
	token := u.data().AuthenticationToken()
	if token == nil {
		return
	}
	return token.LoID.ID, token.LoID.ID != ""
}

func (u User) IsAccountType(accountTypes ...string) (string, bool) {
	token := u.data().AuthenticationToken()
	if token == nil {
		return AccountTypeUnknown, false
	}
	if slices.Contains(accountTypes, token.AccountType) {
		return token.AccountType, true
	}
	return AccountTypeUnknown, false
}

// CookieCreatedAt returns the time the cookie was created.
func (u User) CookieCreatedAt() (ts time.Time, ok bool) {
	if insecure := u.data().InsecureCookieCreatedAt; !insecure.IsZero() {
		return insecure, true
	}
	token := u.data().AuthenticationToken()
	if token == nil {
		return
	}
	ts = token.LoID.CreatedAt.ToTime()
	return ts, !ts.IsZero()
}

// Roles returns the roles the user has.
func (u User) Roles() []string {
	token := u.data().AuthenticationToken()
	if token == nil {
		return nil
	}
	return token.Roles
}

// HasRole returns true if the user has the specific role.
func (u User) HasRole(role string) bool {
	// Since in most cases the roles slice would be quite small,
	// it's better to iterate them than converting the slice into a set.
	for _, r := range u.Roles() {
		if strings.EqualFold(role, r) {
			return true
		}
	}
	return false
}

// UpdateExperimentEvent updates the passed in experiment event with user info.
//
// It always updates UserID, LoggedIn, and CookieCreatedAt fields and never
// touches other fields.
func (u User) UpdateExperimentEvent(ee *experiments.ExperimentEvent) {
	ee.UserID, _ = u.LoID()
	ee.LoggedIn = thrift.BoolPtr(u.IsLoggedIn())
	ee.CookieCreatedAt, _ = u.CookieCreatedAt()
}

// VariantInputs returns the map containing the user related fields that should
// be used in experiments.Variant call.
func (u User) VariantInputs() map[string]interface{} {
	var ee experiments.ExperimentEvent
	u.UpdateExperimentEvent(&ee)

	// Reference for the keys:
	// https://github.com/reddit/edgecontext.py/blob/420e58728ee7085a2f91c5db45df233142b251f9/reddit_edgecontext/__init__.py#L262-L266
	return map[string]interface{}{
		"user_id":                  stringOrNil(ee.UserID),
		"logged_in":                *ee.LoggedIn,
		"cookie_created_timestamp": timebp.TimeToMilliseconds(ee.CookieCreatedAt),
	}
}

func stringOrNil(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
