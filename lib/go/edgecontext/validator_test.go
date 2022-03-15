package edgecontext_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"

	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

// copied from https://github.com/reddit/edgecontext.py/blob/420e58728ee7085a2f91c5db45df233142b251f9/tests/edge_context_tests.py#L54
const validToken = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwfQ.dRzzfc9GmzyqfAbl6n_C55JJueraXk9pp3v0UYXw0ic6W_9RVa7aA1zJWm7slX9lbuYldwUtHvqaSsOpjF34uqr0-yMoRDVpIrbkwwJkNuAE8kbXGYFmXf3Ip25wMHtSXn64y2gJN8TtgAAnzjjGs9yzK9BhHILCDZTtmPbsUepxKmWTiEX2BdurUMZzinbcvcKY4Rb_Fl0pwsmBJFs7nmk5PvTyC6qivCd8ZmMc7dwL47mwy_7ouqdqKyUEdLoTEQ_psuy9REw57PRe00XCHaTSTRDCLmy4gAN6J0J056XoRHLfFcNbtzAmqmtJ_D9HGIIXPKq-KaggwK9I4qLX7g`

func TestValidToken(t *testing.T) {
	token, err := globalTestImpl.ValidateToken(validToken)
	if err != nil {
		t.Fatal(err)
	}
	expected := "t2_example"
	actual := token.Subject()
	if actual != expected {
		t.Errorf("subject expected %q, got %q", expected, actual)
	}
}

const (
	testPubKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtzMnDEQPd75QZByogNlB
NY2auyr4sy8UNTDARs79Edq/Jw5tb7ub412mOB61mVrcuFZW6xfmCRt0ILgoaT66
Tp1RpuEfghD+e7bYZ+Q2pckC1ZaVPIVVf/ZcCZ0tKQHoD8EpyyFINKjCh516VrCx
KuOm2fALPB/xDwDBEdeVJlh5/3HHP2V35scdvDRkvr2qkcvhzoy0+7wUWFRZ2n6H
TFrxMHQoHg0tutAJEkjsMw9xfN7V07c952SHNRZvu80V5EEpnKw/iYKXUjCmoXm8
tpJv5kXH6XPgfvOirSbTfuo+0VGqVIx9gcomzJ0I5WfGTD22dAxDiRT7q7KZnNgt
TwIDAQAB
-----END PUBLIC KEY-----`

	expectedFingerprint = "SHA256:lZ0hkWRsDpapeBu2ekX9WY2oYInHwdRaXTwtBecDicI"
)

func TestFingerprint(t *testing.T) {
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(testPubKeyPEM))
	if err != nil {
		t.Fatalf("Unable to parse pub key from PEM: %v", err)
	}
	fingerprint, err := edgecontext.RSAPublicKeyFingerprint(pubKey)
	if err != nil {
		t.Errorf("Unable to calculate fingerprint from pub key: %v", err)
	}
	if fingerprint != expectedFingerprint {
		t.Errorf("Fingerprint got %q, want %q", fingerprint, expectedFingerprint)
	}
}
