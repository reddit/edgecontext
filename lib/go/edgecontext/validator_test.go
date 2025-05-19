package edgecontext_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/reddit/edgecontext/lib/go/edgecontext"
)

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
