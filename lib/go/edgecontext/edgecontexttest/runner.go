package edgecontexttest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/gofrs/uuid"
	"github.com/reddit/baseplate.go/experiments"
	"github.com/reddit/baseplate.go/secrets"
	"github.com/reddit/baseplate.go/timebp"
	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

const (
	headerWithNoAuth                     = "\f\x00\x01\v\x00\x01\x00\x00\x00\vt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\f\x00\x02\v\x00\x01\x00\x00\x00\bbeefdead\x00\v\x00\x03\x00\x00\x02\x11eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJleHAiOjE3NDc4ODkzODksIm5iZiI6MTc0Nzg4OTI2OSwiaWF0IjoxNzQ3ODg5MzI5fQ.uRnz-pWFp6Fj3vgGnZPDds9pBcWdh2G3vJm1jLMKTEDswlVS0leZ0XSvGzN5JbX9Em7NVL3-Gb0NMtvDkC7TOcIlHWxLD891ooOtDDJwIubUuHnIQr5h4x6DhdmLxu0wYB38v_JrJVHlLywjK_YdXVA-Bw6mgq9inxB6PzB4BrH3Q2upU4IpGsrPNjMk6hD2qUaAwBi7xKWgsu8D3xHxTXQ1VZBAnC3QGEyE1WYABNDzNAYzsmZ5-M6ke3syAAUtAbweUWf6tUQ2TuTY_Nxnl87q1V9a4y27ecB6NVJdzrCDw8yrF2cSSEEWX6Yi6FDxdHJ91bwlXBbZIwAwMI5SZA\f\x00\x04\v\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\f\x00\x05\x00\f\x00\x06\x00\x00"
	headerWithExpiredAuth                = "\f\x00\x01\v\x00\x01\x00\x00\x00\vt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\f\x00\x02\v\x00\x01\x00\x00\x00\bbeefdead\x00\v\x00\x03\x00\x00\x02\x11eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJleHAiOjE3NDc4ODkzODksIm5iZiI6MTc0Nzg4OTI2OSwiaWF0IjoxNzQ3ODg5MzI5fQ.uRnz-pWFp6Fj3vgGnZPDds9pBcWdh2G3vJm1jLMKTEDswlVS0leZ0XSvGzN5JbX9Em7NVL3-Gb0NMtvDkC7TOcIlHWxLD891ooOtDDJwIubUuHnIQr5h4x6DhdmLxu0wYB38v_JrJVHlLywjK_YdXVA-Bw6mgq9inxB6PzB4BrH3Q2upU4IpGsrPNjMk6hD2qUaAwBi7xKWgsu8D3xHxTXQ1VZBAnC3QGEyE1WYABNDzNAYzsmZ5-M6ke3syAAUtAbweUWf6tUQ2TuTY_Nxnl87q1V9a4y27ecB6NVJdzrCDw8yrF2cSSEEWX6Yi6FDxdHJ91bwlXBbZIwAwMI5SZA\f\x00\x04\x00\f\x00\x05\x00\f\x00\x06\x00\x00"
	headerWithAnonAuth                   = "\f\x00\x01\v\x00\x01\x00\x00\x00\vt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\f\x00\x02\v\x00\x01\x00\x00\x00\bbeefdead\x00\v\x00\x03\x00\x00\x02.eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJleHAiOjI1MjQ2MDgwMDAsIm5iZiI6MTc0Nzg4OTI2OSwiaWF0IjoxNzQ3ODg5MzI5LCJyb2xlcyI6WyJhbm9ueW1vdXMiXX0.TmeI4HaJmwfVziCMmodL7M6h_UbBYSv_0_cFjRtq8F9YAxjaOY0s_fXGuuBx1SkFzKpxyV8nAowHQ82xuPRq0X4El4E86otGi87k9yM9Ij5_c5Xjud8_XuH2Hmb8GIMJ35MaPPjtKMSpB9r-kxUktB-P55LhI8z_FfGpKd3lCEhyz1GLSnkLiIP7l72Ertz722YWOpN5jR2R2dftH8_4Om8PC7WnJNG2iGZL6V_9g6csy3e_96C1iiFGx2WSNhdGHIDC1OULaobHNa4JFdwf7IglXLgUboekunyx-4s8pVaL36oNMbGQtAhuwZ_rM0gMV3Hy1wgg_ZOMadBQiDi2Lw\f\x00\x04\x00\f\x00\x05\x00\f\x00\x06\x00\x00"
	headerWithValidServiceAuth           = "\f\x00\x01\x00\f\x00\x02\x00\v\x00\x03\x00\x00\x027eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJzdWIiOiJzZXJ2aWNlL3Rlc3Qtc2VydmljZSIsImV4cCI6MjUyNDYwODAwMCwibmJmIjoxNzQ3ODg5MjY5LCJpYXQiOjE3NDc4ODkzMjl9.KBCbNnT7qxDn7NKXS_sR4pisDKe3xhQ0Hxjuj09zD45U0bM39qTR0n2Dqhp4lPQQccItKkeQ94SFLbK7wO0Z4B0d2mvkuCY_CJMwnhB3EwvD4Qwxg6tstOckr9TAYKRLID8OPLcKafskj7lsO5_w_WoSj9LifjbPKOsDu5fMATRHNHhPcIx4tdPhnPq37elu8TtW-AB-jJePKj6qNjhZDugNATFJoK-cRwhnypU9LmfWCEpHoivgV2GcliYjgwrdVkM4JW-FT9MAco4646puTSaMQ8YS9iW7i_WWIP0TkfzaItxWtQlVl7aykZLZzZTvsQgVi3XHiStHXww46X83ug\f\x00\x04\x00\f\x00\x05\v\x00\x01\x00\x00\x00\x0eorigin/service\x00\f\x00\x06\x00\f\x00\a\v\x00\x01\x00\x00\x00$1566dce9-9567-4952-b23b-9fd72e111162\x00\x00"
	headerWithValidServiceAuthAndOptions = "\f\x00\x01\x00\f\x00\x02\x00\v\x00\x03\x00\x00\x02\x83eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJzdWIiOiJzZXJ2aWNlL3Rlc3Qtc2VydmljZSIsImV4cCI6MjUyNDYwODAwMCwibmJmIjoxNzQ3ODg5MjY5LCJpYXQiOjE3NDc4ODkzMjksIm9ibyI6eyJhaWQiOiJ0Ml9kZWFkYmVlZiIsInJvbGVzIjpbImFkbWluIl19LCJzZWEiOnRydWV9.dje0WVHiN2Greb6Y_97KJ52E10vhh88qbP0X3XFYR3oxIJSxyIt7gite_fCPsbO3Y0jPfMGi5IcYKOsv2n5XjRxlYzUFpy68Qu1ZcPWO3ZiICEi-_wSxJgj4djFqeRArGTuc8AzAsM4IVGuQlUEPWkmxMvLFpsFmpVHwI_srqhhdb04KIniSj4t8I3YY9Lt4hqDNvN3Y1uT2WtZRrOFJTIwbEiwnuJcFdFUqG1jZfCkOosBZ-bwyFz4-G9IGlpt4lPs462GyR_DyTQ1cmnOC0K7Uz8Q9QcOvXa9eACSk6tL93Xi5-x4yxuMFNIBsMBZWuKrnfBWugxwsnTnhzRVRDQ\f\x00\x04\x00\f\x00\x05\v\x00\x01\x00\x00\x00\x0eorigin/service\x00\f\x00\x06\x00\f\x00\a\v\x00\x01\x00\x00\x00$a3b2d5c2-ab27-4948-9dae-78a3ffb46957\x00\x00"
	headerWithValidAuth                  = "\f\x00\x01\v\x00\x01\x00\x00\x00\vt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\f\x00\x02\v\x00\x01\x00\x00\x00\bbeefdead\x00\v\x00\x03\x00\x00\x02*eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwLCJuYmYiOjE3NDc4ODkyNjksImlhdCI6MTc0Nzg4OTMyOX0.sxfTZVoT-IBLlHQQ7frsCXlXyBnqqVueR7ZKh8KjrkDMzUMGSw4xjIXmpKLCASj8OVoVQ8una3kQASM18gcmLcwZXoJEiaYBWyB7TYu-OXL2prX9zJbBWVoPge_iUZ86Qt2ddJE5LxTDpVZ0BkdmDe9jE2j2RE5igyLaydfLdtkbNHDuvxeHnKFhj51qeez5Y3kf6fmJk-dAItKRFunQ_mCDPzvMsAwq884bw2ayh_M9YH1M7KfuYPi8x-vJwMgqVTOIXgLYoPWkQ2ZPfaLmrqfGHzdKoD6dfuOG5jnbZzkXOg3QZ5uPyRMNERkarfmynDaR-YePReFbLpE-1Kpwng\f\x00\x04\v\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\f\x00\x05\v\x00\x01\x00\x00\x00\tbaseplate\x00\f\x00\x06\v\x00\x01\x00\x00\x00\x02OK\x00\f\x00\a\v\x00\x01\x00\x00\x00$2adaff94-9067-4de0-a00b-79fded5cff9e\x00\f\x00\b\v\x00\x01\x00\x00\x00\x05en_US\x00\x00"

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

	validToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjphaGxBYkE5SW9qM1VGTUhzOEc1ZWFia1JBSG1hWlpaTm9OREJZQW40OHBVIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwLCJuYmYiOjE3NDc4ODkyNjksImlhdCI6MTc0Nzg4OTMyOX0.sxfTZVoT-IBLlHQQ7frsCXlXyBnqqVueR7ZKh8KjrkDMzUMGSw4xjIXmpKLCASj8OVoVQ8una3kQASM18gcmLcwZXoJEiaYBWyB7TYu-OXL2prX9zJbBWVoPge_iUZ86Qt2ddJE5LxTDpVZ0BkdmDe9jE2j2RE5igyLaydfLdtkbNHDuvxeHnKFhj51qeez5Y3kf6fmJk-dAItKRFunQ_mCDPzvMsAwq884bw2ayh_M9YH1M7KfuYPi8x-vJwMgqVTOIXgLYoPWkQ2ZPfaLmrqfGHzdKoD6dfuOG5jnbZzkXOg3QZ5uPyRMNERkarfmynDaR-YePReFbLpE-1Kpwng"

	verificationSecret = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAziSyW/ychrelzKu0unCI\npuo3eXokDiywKoUSKlvWOWujJZHWWakIxWuMWZu+L756dVYcgjJfQM/pZo3AlPxy\nuQICiuKH9Ol1tFJV+bdd0IC2uJKSaqZLXslK+cQpPQdnWTmBTR1TXsqSQx3kuRfq\nQjF4R6Yr7mdSIX8/uZg82WIVCR24++G8/ENSNJ+H2Gi/YujO/3Qc5ujPwpwIsTkg\nPQLxotvYbQsRyUZfuA6EQlUZpKbJoBUriNb0+Z+QANTD/hJjUNws/NNSHI6aWfhW\ngMQQSTmd9L5pxyY2m5kzJgMXQKWHlj4PjaKo5TqS+SmN2Vm753oUNFGsBybDNPAv\n2wIDAQAB\n-----END PUBLIC KEY-----\n"
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
	expectedUser        = "t2_example"

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
		t.Errorf("Expected non-nil ExperimentEvent.LoggedIn")
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

func edgeContextWithLocaleCode(l string, impl *edgecontext.Impl) error {
	ctx := context.Background()
	_, err := edgecontext.New(
		ctx,
		impl,
		edgecontext.NewArgs{
			LocaleCode: l,
		},
	)
	return err
}

type runnerConfig struct {
	useDefaultBackend bool
	backend           *edgecontext.Backend
}

// RunnerOption is a function that configures the test Runner.
type RunnerOption func(*runnerConfig)

// WithBackend sets the backend for the Runner.
func WithBackend(backend *edgecontext.Backend) RunnerOption {
	return func(c *runnerConfig) {
		c.backend = backend
	}
}

// WithDefaultBackend configures the Runner to use the default backend.
func WithDefaultBackend() RunnerOption {
	return func(c *runnerConfig) {
		c.useDefaultBackend = true
	}
}

// Runner is a test runner for edgecontext package. This can be used to run tests to verify that a new Backend works.
type Runner struct {
	impl *edgecontext.Impl
}

// NewRunner creates a new Runner instance for testing. It requires either a backend to be set or the use of WithDefaultBackend.
func NewRunner(t *testing.T, opts ...RunnerOption) *Runner {
	var cfg runnerConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.backend != nil && !cfg.useDefaultBackend {
		edgecontext.SetBackend(cfg.backend)
		t.Cleanup(func() {
			edgecontext.SetBackend(edgecontext.V0Backend())
		})
	} else if cfg.backend == nil && cfg.useDefaultBackend {
		// nothing to do here, this is just to simplify the configuration
		// check code and make it clear what cases are valid.
	} else {
		t.Fatalf("must specify either backend or use WithDefaultBackend")
	}

	store, _, err := secrets.NewTestSecrets(
		context.Background(),
		map[string]secrets.GenericSecret{
			"secret/authentication/public-key": {
				Type:     secrets.VersionedType,
				Encoding: secrets.IdentityEncoding,
				Current:  verificationSecret,
			},
		},
	)

	if err != nil {
		t.Fatalf("creating test secrets: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	return &Runner{
		impl: edgecontext.Init(edgecontext.Config{Store: store}),
	}
}

// Run runs all the tests defined in the Runner.
func (r *Runner) Run(t *testing.T) {
	for name, fn := range map[string]func(*testing.T){
		"TestNew":        r.TestNew,
		"TestLocale":     r.TestLocale,
		"TestFromHeader": r.TestFromHeader,
	} {
		t.Run(name, fn)
	}
}

// TestNew tests the creation of a new EdgeContext with valid parameters.
func (r *Runner) TestNew(t *testing.T) {
	ctx := context.Background()
	e, err := edgecontext.New(
		ctx,
		r.impl,
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
		t.Fatalf("initializing edge context: %v", err)
	}
	if e.Header() != headerWithValidAuth {
		t.Errorf("Header mismatch:\nwant: %q\ngot:  %q", headerWithValidAuth, e.Header())
	}

	if diff := Diff(Want(
		WithAuthenticationToken(
			WithLoggedInUser(expectedUser, expectedCookieTime, []string{}),
		),
		WithRequestMetadata(expectedRequestID, expectedSessionID, expectedDeviceID),
		WithLocation(expectedLocaleCode, expectedCountryCode),
		WithOriginService(expectedOrigin),
		WithLoIDCookie(expectedLoID, expectedCookieTime),
	), e); diff != "" {
		t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
	}
}

// TestLocale tests the validation of locale codes in EdgeContext.
func (r *Runner) TestLocale(t *testing.T) {
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
			err := edgeContextWithLocaleCode(c.locale, r.impl)

			if c.valid {
				if err != nil {
					t.Errorf("Did not expect error, got %v", err)
				}
			} else {
				if !errors.Is(err, edgecontext.ErrInvalidLocaleCode) {
					t.Errorf("Expected "+
						"ErrInvalidLocaleCode, got %v", err)
				}
			}
		})
	}
}

// TestFromHeader tests the creation of an EdgeContext from a header string.
func (r *Runner) TestFromHeader(t *testing.T) {

	t.Run(
		"no-header",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), "", r.impl)
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
			e, err := edgecontext.FromHeader(context.Background(), headerWithNoAuth, r.impl)
			if err != nil {
				t.Fatal(err)
			}

			if diff := Diff(Want(
				WithNoAuthenticationToken(),
				WithRequestMetadata("", expectedSessionID, expectedDeviceID),
				WithLoIDCookie(expectedLoID, expectedCookieTime),
			), e); diff != "" {
				t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
			}

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
		},
	)

	t.Run(
		"valid-auth",
		func(t *testing.T) {
			e, err := edgecontext.FromHeader(context.Background(), headerWithValidAuth, r.impl)
			if err != nil {
				t.Fatal(err)
			}

			if diff := Diff(Want(
				WithAuthenticationToken(
					WithLoggedInUser(expectedUser, expectedCookieTime, []string{}),
				),
				WithRequestMetadata(expectedRequestID, expectedSessionID, expectedDeviceID),
				WithLocation(expectedLocaleCode, expectedCountryCode),
				WithOriginService(expectedOrigin),
			), e); diff != "" {
				t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
			}

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
			e, err := edgecontext.FromHeader(context.Background(), headerWithExpiredAuth, r.impl)
			if err != nil {
				t.Fatal(err)
			}

			if diff := Diff(Want(
				WithNoAuthenticationToken(),
				WithRequestMetadata("", expectedSessionID, ""),
				WithLoIDCookie(expectedLoID, expectedCookieTime),
			), e); diff != "" {
				t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
			}

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
			e, err := edgecontext.FromHeader(context.Background(), headerWithAnonAuth, r.impl)
			if err != nil {
				t.Fatal(err)
			}

			if diff := Diff(Want(
				WithAuthenticationToken(
					WithLoggedOutUser(expectedUser, expectedCookieTime, []string{"anonymous"}),
				),
				WithRequestMetadata("", expectedSessionID, ""),
				WithLoIDCookie(expectedLoID, expectedCookieTime),
			), e); diff != "" {
				t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
			}

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
			e, err := edgecontext.FromHeader(context.Background(), headerWithReadableRequestID, r.impl)
			if err != nil {
				t.Fatal(err)
			}

			if diff := Diff(Want(
				WithNoAuthenticationToken(),
				WithRequestMetadata(expectedRequestID, "", ""),
			), e); diff != "" {
				t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
			}
		},
	)

	t.Run(
		"service",
		func(t *testing.T) {
			t.Run("service auth only", func(t *testing.T) {
				e, err := edgecontext.FromHeader(context.Background(), headerWithValidServiceAuth, r.impl)
				if err != nil {
					t.Fatal(err)
				}

				if diff := Diff(Want(
					WithAuthenticationToken(
						WithService(expectedServiceName),
					),
					WithRequestMetadata("1566dce9-9567-4952-b23b-9fd72e111162", "", ""),
					WithOriginService("origin/service"),
				), e); diff != "" {
					t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
				}
			})

			t.Run("with additional options", func(t *testing.T) {
				e, err := edgecontext.FromHeader(context.Background(), headerWithValidServiceAuthAndOptions, r.impl)
				if err != nil {
					t.Fatal(err)
				}

				if diff := Diff(Want(
					WithAuthenticationToken(
						WithService(expectedServiceName,
							WithOnBehalfOfUser(expectedLoID, []string{"admin"}),
							WithServiceRequestsElevatedAccess(),
						),
					),
					WithRequestMetadata("a3b2d5c2-ab27-4948-9dae-78a3ffb46957", "", ""),
					WithOriginService("origin/service"),
				), e); diff != "" {
					t.Fatalf("EdgeRequestContext mismatch (-want +got):\n%s", diff)
				}
			})
		},
	)
}
