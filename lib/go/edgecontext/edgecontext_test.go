package edgecontext_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/reddit/baseplate.go/detach"
	"github.com/reddit/baseplate.go/experiments"
	"github.com/reddit/baseplate.go/timebp"

	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

const (
	// copied from https://github.com/reddit/edgecontext.py/blob/420e58728ee7085a2f91c5db45df233142b251f9/tests/edge_context_tests.py#L55-L58
	headerWithNoAuth            = "\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0c\x00\x04\x0b\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\x00"
	headerWithValidAuth         = "\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xaeeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwfQ.dRzzfc9GmzyqfAbl6n_C55JJueraXk9pp3v0UYXw0ic6W_9RVa7aA1zJWm7slX9lbuYldwUtHvqaSsOpjF34uqr0-yMoRDVpIrbkwwJkNuAE8kbXGYFmXf3Ip25wMHtSXn64y2gJN8TtgAAnzjjGs9yzK9BhHILCDZTtmPbsUepxKmWTiEX2BdurUMZzinbcvcKY4Rb_Fl0pwsmBJFs7nmk5PvTyC6qivCd8ZmMc7dwL47mwy_7ouqdqKyUEdLoTEQ_psuy9REw57PRe00XCHaTSTRDCLmy4gAN6J0J056XoRHLfFcNbtzAmqmtJ_D9HGIIXPKq-KaggwK9I4qLX7g\x0c\x00\x04\x0b\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\x0c\x00\x05\x0b\x00\x01\x00\x00\x00\tbaseplate\x00\x0c\x00\x06\x0b\x00\x01\x00\x00\x00\x02OK\x00\x0c\x00\x07\x0b\x00\x01\x00\x00\x00\x24" + expectedRequestID + "\x00\f\x00\b\v\x00\x01\x00\x00\x00\x05en_US\x00\x00"
	headerWithExpiredAuth       = "\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xaeeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoxMjYyMzA0MDAwfQ.iUD0J2blW-HGtH86s66msBXymCRCgyxAZJ6xX2_SXD-kegm-KjOlIemMWFZtsNv9DJI147cNP81_gssewvUnhIHLVvXWCTOROasXbA9Yf2GUsjxoGSB7474ziPOZquAJKo8ikERlhOOVk3r4xZIIYCuc4vGZ7NfqFxjDGKAWj5Tt4VUiWXK1AdxQck24GyNOSXs677vIJnoD8EkgWqNuuwY-iFOAPVcoHmEuzhU_yUeQnY8D-VztJkip5-YPEnuuf-dTSmPbdm9ZTOP8gjTsG0Sdvb9NdLId0nEwawRy8CfFEGQulqHgd1bqTm25U-NyXQi7zroi1GEdykZ3w9fVNQ\x0c\x00\x07\x00\x00"
	headerWithAnonAuth          = "\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xc0eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlcyI6WyJhbm9ueW1vdXMiXSwic3ViIjpudWxsLCJleHAiOjI1MjQ2MDgwMDB9.gQDiVzOUh70mKKK-YBTnLHWBOEuQyRllEE1-EIMfy3x5K8PsH9FB6Oy9S5HbILjfGFNrIBeux9HyW6hBDikoZDhn5QWyPNitL1pzMNONGGrXzSfaDoDbFy4MLD03A7zjG3qWBn_wLjgzUXX6qVX6W_gWO7dMqrq0iFvEegue-xQ1HGiXfPgnTrXRRovUO3JHy1LcZsmOjltYj5VGUTWXodBM8ObKEealDxg8yskEPy0IuujNMmb9eIyuHB8Ozzpg-lr790lxP37s5HCf18vrZ-IhRmLcLCqm5WSFyq_Ld2ByblBKL9pPst1AZYZTXNRIqovTAqr6v0-xjUeJ1iho9A\x0c\x00\x07\x00\x00"
	headerWithReadableRequestID = ("\x0c\x00\x01\x00\x0c\x00\x02\x00\x0c\x00\x04\x00\x0c\x00\x05\x00\x0c\x00\x06\x00" +
		// struct 7: request_id
		"\x0c\x00\x07" +
		// string 1: readable_id
		"\x0b\x00\x01" +
		// length == 36, payload
		"\x00\x00\x00\x24" + expectedRequestID +
		// end of struct
		"\x00" +
		// end of struct
		"\x00")

	headerWithValidServiceAuth           = "\x0c\x00\x01\x00\x0c\x00\x02\x00\x0b\x00\x03\x00\x00\x02\x0aeyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpsWjBoa1dSc0RwYXBlQnUyZWtYOVdZMm9ZSW5Id2RSYVhUd3RCZWNEaWNJIiwidHlwIjoiSldUIn0.eyJzdWIiOiJzZXJ2aWNlL3Rlc3Qtc2VydmljZSIsImV4cCI6MjUyNDYwODAwMH0.P41Iahxu-Bbg5srTFSQTBkzwiff4ytlhVBUYuyYTFGY_7XCyKdZywUmVHRY_Q2w8Q2uaybnmuoM95JhRpdNYcTPIYWEby4Z5DSV-zMqqmHnP22aH_sAckFQl86Yw_2pdZpKKJ-KQkyT0vEkxe-vNs5HhEdBr6Rae0g2SKEr7RaPMoToq6xpucDAREVWa7yJMtyyNtiVixeLoxTegRLOZTFEVt4TTYKDuT2FdEY5P2b8BOSpFMoiv9w51gZO1qvn9Zjrl00Z-lI_onihMIkrG_viWVAlzEl8d5ZWuJVjHJvm7O0CS4OuhZocE2qbYQrw9THSS1Mh4YR-_r2v1ArYnVA\x0c\x00\x04\x00\x0c\x00\x05\x0b\x00\x01\x00\x00\x00\x0eorigin/service\x00\x0c\x00\x06\x00\x0c\x00\x07\x0b\x00\x01\x00\x00\x00$1566dce9-9567-4952-b23b-9fd72e111162\x00\x00"
	headerWithValidServiceAuthAndOptions = "\x0c\x00\x01\x00\x0c\x00\x02\x00\x0b\x00\x03\x00\x00\x02VeyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpsWjBoa1dSc0RwYXBlQnUyZWtYOVdZMm9ZSW5Id2RSYVhUd3RCZWNEaWNJIiwidHlwIjoiSldUIn0.eyJzdWIiOiJzZXJ2aWNlL3Rlc3Qtc2VydmljZSIsImV4cCI6MjUyNDYwODAwMCwib2JvIjp7ImFpZCI6InQyX2RlYWRiZWVmIiwicm9sZXMiOlsiYWRtaW4iXX0sInNlYSI6dHJ1ZX0.PVefAKWUFfk_7QKen6Iz0Cfu95Yp92lYETlrxCUacLsa9u-qz36aet21iwFrdnJiz7gDeJRH7sOJyh6jRmkD0ptWs4Zl7VqpZY-ALgDOdhwSHoUIoV2L7twT-Dm3Tdyfbzq01fOni9ioq5akKnETC5IbLSOqp1ssWJcgo_9g-X-SdRiuf5u8YHD2Mrep5U21bkbYnm4rK9tX_oCnhrrp4rbXi5yogx594oNmOWUedIeyv6QY_xVGbaXOz7deBIWQY2fSYG3cpiBNtSYEJ4yDTbjGY0G1Vp78bX8YZlboc13TGoDpARdfHuHeQU0wAQEhi7pu0Q4FufEVua4q1f0P3A\x0c\x00\x04\x00\x0c\x00\x05\x0b\x00\x01\x00\x00\x00\x0eorigin/service\x00\x0c\x00\x06\x00\x0c\x00\x07\x0b\x00\x01\x00\x00\x00$a3b2d5c2-ab27-4948-9dae-78a3ffb46957\x00\x00"
)

const (
	expectedCountryCode = "OK"
	expectedLocaleCode  = "en_US"
	expectedDeviceID    = "becc50f6-ff3d-407a-aa49-fa49531363be"
	expectedLoID        = "t2_deadbeef"
	expectedOrigin      = "baseplate"
	expectedSessionID   = "beefdead"
	expectedRequestID   = "2adaff94-9067-4de0-a00b-79fded5cff9e"
	expectedServiceName = "test-service"

	emptyDeviceID = "00000000-0000-0000-0000-000000000000"
)

var expectedCookieTime = time.Unix(100, 0)

var uuidGen = uuid.NewGen()

func mustV4() uuid.UUID {
	id, err := uuidGen.NewV4()
	if err != nil {
		panic(err)
	}
	return id
}

var emptyExperimentEventBase = experiments.ExperimentEvent{}

var fullExperimentEventBase = experiments.ExperimentEvent{
	ID:            mustV4(),
	CorrelationID: mustV4(),
	DeviceID:      mustV4(),
	Experiment: &experiments.ExperimentConfig{
		ID:             1234,
		Name:           "name",
		Owner:          "owner",
		Enabled:        thrift.BoolPtr(true),
		Version:        "version",
		Type:           "type",
		StartTimestamp: timebp.TimestampSecondF(time.Now()),
		StopTimestamp:  timebp.TimestampSecondF(time.Now()),
	},
	VariantName:     "variant",
	UserID:          "t2_user",
	LoggedIn:        thrift.BoolPtr(true),
	CookieCreatedAt: time.Now(),
	OAuthClientID:   "client",
	ClientTimestamp: time.Now(),
	AppName:         "app",
	SessionID:       "session",
	IsOverride:      true,
	EventType:       "type",
}

func compareUntouchedFields(t *testing.T, expected, actual experiments.ExperimentEvent) {
	t.Helper()

	if expected.ID.String() != actual.ID.String() {
		t.Errorf(
			"Expected ExperimentEvent.ID %v, got %v",
			expected.ID,
			actual.ID,
		)
	}

	if expected.CorrelationID.String() != actual.CorrelationID.String() {
		t.Errorf(
			"Expected ExperimentEvent.CorrelationID %v, got %v",
			expected.CorrelationID,
			actual.CorrelationID,
		)
	}

	if expected.Experiment != actual.Experiment {
		t.Errorf(
			"Expected ExperimentEvent.Experiment %#v, got %#v",
			expected.Experiment,
			actual.Experiment,
		)
	}

	if expected.VariantName != actual.VariantName {
		t.Errorf(
			"Expected ExperimentEvent.VariantName %v, got %v",
			expected.VariantName,
			actual.VariantName,
		)
	}

	if !expected.ClientTimestamp.Equal(actual.ClientTimestamp) {
		t.Errorf(
			"Expected ExperimentEvent.ClientTimestamp %v, got %v",
			expected.ClientTimestamp,
			actual.ClientTimestamp,
		)
	}

	if expected.AppName != actual.AppName {
		t.Errorf(
			"Expected ExperimentEvent.AppName %v, got %v",
			expected.AppName,
			actual.AppName,
		)
	}

	if expected.IsOverride != actual.IsOverride {
		t.Errorf(
			"Expected ExperimentEvent.IsOverride %v, got %v",
			expected.IsOverride,
			actual.IsOverride,
		)
	}

	if expected.EventType != actual.EventType {
		t.Errorf(
			"Expected ExperimentEvent.EventType %v, got %v",
			expected.EventType,
			actual.EventType,
		)
	}
}

func compareTouchedFields(
	t *testing.T,
	actual experiments.ExperimentEvent,
	userID string,
	loggedIn bool,
	cookieTime time.Time,
	oauthID string,
	session string,
	device string,
) {
	t.Helper()

	if userID != actual.UserID {
		t.Errorf(
			"Expected ExperimentEvent.UserID %v, got %v",
			userID,
			actual.UserID,
		)
	}

	if actual.LoggedIn == nil {
		t.Error("Expected non-nil ExperimentEvent.LoggedIn")
	} else {
		if loggedIn != *actual.LoggedIn {
			t.Errorf(
				"Expected ExperimentEvent.LoggedIn %v, got %v",
				loggedIn,
				*actual.LoggedIn,
			)
		}
	}

	if !cookieTime.Equal(actual.CookieCreatedAt) {
		t.Errorf(
			"Expected ExperimentEvent.CookieCreatedAt %v, got %v",
			cookieTime,
			actual.CookieCreatedAt,
		)
	}

	if oauthID != actual.OAuthClientID {
		t.Errorf(
			"Expected ExperimentEvent.OAuthClientID %v, got %v",
			oauthID,
			actual.OAuthClientID,
		)
	}

	if session != actual.SessionID {
		t.Errorf(
			"Expected ExperimentEvent.SessionID %v, got %v",
			session,
			actual.SessionID,
		)
	}

	if device != actual.DeviceID.String() {
		t.Errorf(
			"Expected ExperimentEvent.DeviceID %v, got %v",
			device,
			actual.DeviceID,
		)
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	e, err := edgecontext.New(
		ctx,
		globalTestImpl,
		edgecontext.NewArgs{
			LoID:              expectedLoID,
			LoIDCreatedAt:     expectedCookieTime,
			SessionID:         expectedSessionID,
			AuthToken:         validToken,
			DeviceID:          expectedDeviceID,
			CountryCode:       expectedCountryCode,
			OriginServiceName: expectedOrigin,
			RequestID:         expectedRequestID,
			LocaleCode:        expectedLocaleCode,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if e.Header() != headerWithValidAuth {
		t.Errorf("Header expected %q, got %q", headerWithValidAuth, e.Header())
	}
}

func TestLocale(t *testing.T) {
	for _, c := range []struct {
		label  string
		locale string
		valid  bool
	}{
		{
			label:  "valid-language",
			locale: "es",
			valid:  true,
		},
		{
			label:  "valid-language-valid-region",
			locale: "es_MX",
			valid:  true,
		},
		{
			label:  "valid-language-valid-region-hyphen",
			locale: "es-MX",
			valid:  true,
		},
		{
			label:  "invalid-separator",
			locale: "esMX",
			valid:  false,
		},
		{
			label:  "invalid-capitalization",
			locale: "ES_MX",
			valid:  false,
		},
		{
			label:  "valid-alphabet",
			locale: "sr-Latn",
			valid:  true,
		},
		{
			label:  "valid-orthography",
			locale: "de-DE-1996",
			valid:  true,
		},
		{
			label:  "valid-number",
			locale: "es-419",
			valid:  true,
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			err := edgeContextWithLocaleCode(c.locale)

			if c.valid {
				if err != nil {
					t.Errorf("Did not expect error, got %v", err)
				}
			} else {
				if !errors.Is(err, edgecontext.ErrInvalidLocaleCode) {
					t.Errorf("Expected edgecontext.ErrInvalidLocaleCode, got %v", err)
				}
			}
		})
	}
}

func edgeContextWithLocaleCode(l string) error {
	ctx := context.Background()
	_, err := edgecontext.New(
		ctx,
		globalTestImpl,
		edgecontext.NewArgs{
			LocaleCode: l,
		},
	)
	return err
}

func TestFromHeader(t *testing.T) {
	const expectedUser = "t2_example"

	t.Run(
		"no-header",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), "", globalTestImpl)
			if err != nil {
				t.Error(err)
			}
			if e != nil {
				t.Errorf("Expected EdgeRequestContext to be nil, got %#v", e)
			}
		},
	)

	t.Run(
		"no-auth",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithNoAuth, globalTestImpl)
			if err != nil {
				t.Fatal(err)
			}

			if e.SessionID() != expectedSessionID {
				t.Errorf(
					"Expected session id %q, got %q",
					expectedSessionID,
					e.SessionID(),
				)
			}

			if e.DeviceID() != expectedDeviceID {
				t.Errorf(
					"Expected device id %q, got %q",
					expectedDeviceID,
					e.DeviceID(),
				)
			}

			t.Run(
				"user",
				func(t *testing.T) {
					user := e.User()
					if user.IsLoggedIn() {
						t.Error("Expected logged out user, IsLoggedIn() returned true")
					}
					if loid, ok := user.LoID(); !ok {
						t.Error("Failed to get loid from user")
					} else {
						if loid != expectedLoID {
							t.Errorf("LoID expected %q, got %q", expectedLoID, loid)
						}
					}
					if ts, ok := user.CookieCreatedAt(); !ok {
						t.Error("Failed to get cookie created time from user")
					} else {
						if !expectedCookieTime.Equal(ts) {
							t.Errorf(
								"Expected cookie create timestamp %v, got %v",
								expectedCookieTime,
								ts,
							)
						}
					}
				},
			)

			t.Run(
				"auth-token",
				func(t *testing.T) {
					token := e.AuthToken()
					if token != nil {
						t.Errorf("Expected nil auth token, got %+v", *token)
					}
				},
			)

			t.Run(
				"experiment-event",
				func(t *testing.T) {
					// Make deep copy from base
					emptyEvent := emptyExperimentEventBase
					e.UpdateExperimentEvent(&emptyEvent)
					compareUntouchedFields(t, emptyExperimentEventBase, emptyEvent)
					compareTouchedFields(
						t,
						emptyEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						expectedDeviceID,
					)

					// Make deep copy from base
					fullEvent := fullExperimentEventBase
					e.UpdateExperimentEvent(&fullEvent)
					compareUntouchedFields(t, fullExperimentEventBase, fullEvent)
					compareTouchedFields(
						t,
						fullEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						expectedDeviceID,
					)
				},
			)

			t.Run(
				"service",
				func(t *testing.T) {
					_, ok := e.Service()
					if ok {
						t.Errorf("Expected service to be false, got true")
					}
				},
			)
		},
	)

	t.Run(
		"valid-auth",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithValidAuth, globalTestImpl)
			if err != nil {
				t.Fatal(err)
			}

			if e.SessionID() != expectedSessionID {
				t.Errorf(
					"Expected session id %q, got %q",
					expectedSessionID,
					e.SessionID(),
				)
			}

			if e.DeviceID() != expectedDeviceID {
				t.Errorf(
					"Expected device id %q, got %q",
					expectedDeviceID,
					e.DeviceID(),
				)
			}

			if e.CountryCode() != expectedCountryCode {
				t.Errorf(
					"Expected device id %q, got %q",
					expectedCountryCode,
					e.CountryCode(),
				)
			}

			if e.LocaleCode() != expectedLocaleCode {
				t.Errorf(
					"Expected locale code %q, got %q",
					expectedLocaleCode,
					e.LocaleCode(),
				)
			}

			if e.OriginService().Name() != expectedOrigin {
				t.Errorf(
					"Expected origin service %q, got %q",
					expectedOrigin,
					e.OriginService().Name(),
				)
			}

			t.Run(
				"user",
				func(t *testing.T) {
					user := e.User()
					if !user.IsLoggedIn() {
						t.Error("Expected logged in user, IsLoggedIn() returned false")
					}
					if user, ok := user.ID(); !ok {
						t.Error("Failed to get user id")
					} else {
						if user != expectedUser {
							t.Errorf("Expected user id %q, got %q", expectedUser, user)
						}
					}
					if loid, ok := user.LoID(); !ok {
						t.Error("Failed to get loid from user")
					} else {
						// For logged in user,
						// we expect LoID() to return user id instead of loid.
						if loid != expectedUser {
							t.Errorf("LoID expected %q, got %q", expectedUser, loid)
						}
					}
					if ts, ok := user.CookieCreatedAt(); !ok {
						t.Error("Failed to get cookie created time from user")
					} else {
						if !expectedCookieTime.Equal(ts) {
							t.Errorf(
								"Expected cookie create timestamp %v, got %v",
								expectedCookieTime,
								ts,
							)
						}
					}
					if len(user.Roles()) != 0 {
						t.Errorf("Expected empty roles, got %+v", user.Roles())
					}
				},
			)

			t.Run(
				"experiment-event",
				func(t *testing.T) {
					// Make deep copy from base
					emptyEvent := emptyExperimentEventBase
					e.UpdateExperimentEvent(&emptyEvent)
					compareUntouchedFields(t, emptyExperimentEventBase, emptyEvent)
					compareTouchedFields(
						t,
						emptyEvent,
						expectedUser,
						true, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						expectedDeviceID,
					)

					// Make deep copy from base
					fullEvent := fullExperimentEventBase
					e.UpdateExperimentEvent(&fullEvent)
					compareUntouchedFields(t, fullExperimentEventBase, fullEvent)
					compareTouchedFields(
						t,
						fullEvent,
						expectedUser,
						true, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						expectedDeviceID,
					)
				},
			)
		},
	)

	t.Run(
		"expired-auth",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithExpiredAuth, globalTestImpl)
			if err != nil {
				t.Fatal(err)
			}

			if e.SessionID() != expectedSessionID {
				t.Errorf(
					"Expected session id %q, got %q",
					expectedSessionID,
					e.SessionID(),
				)
			}

			t.Run(
				"user",
				func(t *testing.T) {
					user := e.User()
					if user.IsLoggedIn() {
						t.Error("Expected logged out user, IsLoggedIn() returned true")
					}
					if loid, ok := user.LoID(); !ok {
						t.Error("Failed to get loid from user")
					} else {
						if loid != expectedLoID {
							t.Errorf("LoID expected %q, got %q", expectedLoID, loid)
						}
					}
					if ts, ok := user.CookieCreatedAt(); !ok {
						t.Error("Failed to get cookie created time from user")
					} else {
						if !expectedCookieTime.Equal(ts) {
							t.Errorf(
								"Expected cookie create timestamp %v, got %v",
								expectedCookieTime,
								ts,
							)
						}
					}
				},
			)

			t.Run(
				"auth-token",
				func(t *testing.T) {
					token := e.AuthToken()
					if token != nil {
						t.Errorf("Expected nil auth token, got %+v", *token)
					}
				},
			)

			t.Run(
				"experiment-event",
				func(t *testing.T) {
					// Make deep copy from base
					emptyEvent := emptyExperimentEventBase
					e.UpdateExperimentEvent(&emptyEvent)
					compareUntouchedFields(t, emptyExperimentEventBase, emptyEvent)

					// Make deep copy from base
					fullEvent := fullExperimentEventBase
					e.UpdateExperimentEvent(&fullEvent)
					compareUntouchedFields(t, fullExperimentEventBase, fullEvent)
				},
			)

			t.Run(
				"experiment-event",
				func(t *testing.T) {
					// Make deep copy from base
					emptyEvent := emptyExperimentEventBase
					e.UpdateExperimentEvent(&emptyEvent)
					compareUntouchedFields(t, emptyExperimentEventBase, emptyEvent)
					compareTouchedFields(
						t,
						emptyEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						emptyDeviceID,
					)

					// Make deep copy from base
					fullEvent := fullExperimentEventBase
					e.UpdateExperimentEvent(&fullEvent)
					compareUntouchedFields(t, fullExperimentEventBase, fullEvent)
					compareTouchedFields(
						t,
						fullEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						emptyDeviceID,
					)
				},
			)
		},
	)

	t.Run(
		"anon-auth",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithAnonAuth, globalTestImpl)
			if err != nil {
				t.Fatal(err)
			}

			if e.SessionID() != expectedSessionID {
				t.Errorf(
					"Expected session id %q, got %q",
					expectedSessionID,
					e.SessionID(),
				)
			}

			t.Run(
				"user",
				func(t *testing.T) {
					user := e.User()
					if user.IsLoggedIn() {
						t.Error("Expected logged out user, IsLoggedIn() returned true")
					}
					if loid, ok := user.LoID(); !ok {
						t.Error("Failed to get loid from user")
					} else {
						if loid != expectedLoID {
							t.Errorf("LoID expected %q, got %q", expectedLoID, loid)
						}
					}
					if ts, ok := user.CookieCreatedAt(); !ok {
						t.Error("Failed to get cookie created time from user")
					} else {
						if !expectedCookieTime.Equal(ts) {
							t.Errorf(
								"Expected cookie create timestamp %v, got %v",
								expectedCookieTime,
								ts,
							)
						}
					}
					if !user.HasRole("anonymous") {
						t.Errorf(
							"Expected user to have anonymous role, got %+v",
							user.Roles(),
						)
					}
				},
			)

			t.Run(
				"experiment-event",
				func(t *testing.T) {
					// Make deep copy from base
					emptyEvent := emptyExperimentEventBase
					e.UpdateExperimentEvent(&emptyEvent)
					compareUntouchedFields(t, emptyExperimentEventBase, emptyEvent)
					compareTouchedFields(
						t,
						emptyEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						emptyDeviceID,
					)

					// Make deep copy from base
					fullEvent := fullExperimentEventBase
					e.UpdateExperimentEvent(&fullEvent)
					compareUntouchedFields(t, fullExperimentEventBase, fullEvent)
					compareTouchedFields(
						t,
						fullEvent,
						expectedLoID,
						false, // logged in
						expectedCookieTime,
						"", // oauth client id
						expectedSessionID,
						emptyDeviceID,
					)
				},
			)
		},
	)

	t.Run(
		"request-id",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithReadableRequestID, globalTestImpl)
			if err != nil {
				t.Fatal(err)
			}

			if e.RequestID() != expectedRequestID {
				t.Errorf(
					"Expected request id %q, got %q",
					expectedRequestID,
					e.RequestID(),
				)
			}
		},
	)

	t.Run(
		"service",
		func(t *testing.T) {
			t.Run("service auth only", func(t *testing.T) {
				e, err := edgecontext.FromHeader(context.Background(), headerWithValidServiceAuth, globalTestImpl)
				if err != nil {
					t.Fatal(err)
				}

				svc, ok := e.Service()
				if !ok {
					t.Errorf("Expected service to be true, got false")
				}
				if name, ok := svc.Name(); !ok {
					t.Error("Failed to get service name")
				} else {
					if name != expectedServiceName {
						t.Errorf("Expected service name %q, got %q", expectedServiceName, name)
					}
				}

				if id, ok := svc.OnBehalfOfID(); ok {
					t.Errorf("expected no id, got %q", id)
				}

				if roles, ok := svc.OnBehalfOfRoles(); ok {
					t.Errorf("expected no roles, got %q", roles)
				}

				if svc.IsElevatedAccess() {
					t.Errorf("expected no elevated access, got true")
				}
			})

			t.Run("with additional options", func(t *testing.T) {
				e, err := edgecontext.FromHeader(context.Background(), headerWithValidServiceAuthAndOptions, globalTestImpl)
				if err != nil {
					t.Fatal(err)
				}

				svc, ok := e.Service()
				if !ok {
					t.Errorf("Expected service to be true, got false")
				}
				if name, ok := svc.Name(); !ok {
					t.Error("Failed to get service name")
				} else {
					if name != expectedServiceName {
						t.Errorf("Expected service name %q, got %q", expectedServiceName, name)
					}
				}

				if id, ok := svc.OnBehalfOfID(); !ok {
					t.Error("Failed to get on behalf of id")
				} else {
					if id != expectedLoID {
						t.Errorf("Expected on behalf of id %q, got %q", expectedLoID, id)
					}
				}

				if roles, ok := svc.OnBehalfOfRoles(); !ok {
					t.Error("Failed to get on behalf of roles")
				} else {
					if diff := cmp.Diff([]string{"admin"}, roles); diff != "" {
						t.Errorf("mismatch (-want +got)\n%s\n", diff)
					}
				}

				if !svc.IsElevatedAccess() {
					t.Errorf("expected elevated access, got false")
				}
			})
		},
	)
}

func TestDetachIntegration(t *testing.T) {
	ctx := context.Background()
	e, err := edgecontext.New(
		ctx,
		globalTestImpl,
		edgecontext.NewArgs{
			LoID:              expectedLoID,
			LoIDCreatedAt:     expectedCookieTime,
			SessionID:         expectedSessionID,
			AuthToken:         validToken,
			DeviceID:          expectedDeviceID,
			CountryCode:       expectedCountryCode,
			OriginServiceName: expectedOrigin,
			RequestID:         expectedRequestID,
			LocaleCode:        expectedLocaleCode,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx = edgecontext.SetEdgeContext(ctx, e)

	t.Run("inline", func(t *testing.T) {
		inlineCtx, cancel := detach.Inline(ctx, time.Second)
		defer cancel()

		if ec, ok := edgecontext.GetEdgeContext(inlineCtx); ok {
			if ec.Header() != e.Header() {
				t.Fatalf("headers on edge contexts did not match\n\texpected: %q\n\tgot: %q", e.Header(), ec.Header())
			}
		} else {
			t.Fatal("edge context was not set on detached context")
		}
	})

	t.Run("async", func(t *testing.T) {
		detach.Async(ctx, func(asyncCtx context.Context) {
			if ec, ok := edgecontext.GetEdgeContext(asyncCtx); ok {
				if ec.Header() != e.Header() {
					t.Fatalf("headers on edge contexts did not match\n\texpected: %q\n\tgot: %q", e.Header(), ec.Header())
				}
			} else {
				t.Fatal("edge context was not set on detached context")
			}
		})
	})
}
