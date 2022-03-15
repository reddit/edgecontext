package edgecontext

import (
	"context"
	"testing"

	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
)

const (
	validKey1 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtzMnDEQPd75QZByogNlB
NY2auyr4sy8UNTDARs79Edq/Jw5tb7ub412mOB61mVrcuFZW6xfmCRt0ILgoaT66
Tp1RpuEfghD+e7bYZ+Q2pckC1ZaVPIVVf/ZcCZ0tKQHoD8EpyyFINKjCh516VrCx
KuOm2fALPB/xDwDBEdeVJlh5/3HHP2V35scdvDRkvr2qkcvhzoy0+7wUWFRZ2n6H
TFrxMHQoHg0tutAJEkjsMw9xfN7V07c952SHNRZvu80V5EEpnKw/iYKXUjCmoXm8
tpJv5kXH6XPgfvOirSbTfuo+0VGqVIx9gcomzJ0I5WfGTD22dAxDiRT7q7KZnNgt
TwIDAQAB
-----END PUBLIC KEY-----`

	fingerprint1 = "SHA256:lZ0hkWRsDpapeBu2ekX9WY2oYInHwdRaXTwtBecDicI"

	validKey2 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAycU1W/hMRWNLkaJPEwWg
j36URuSaRTV0BEvY+L0nRseCnEdlIsj8LCI+ydk3HlJqj3QicuCP9U0W5JAP4PYB
Xs+dV/J38fqdYfI1myXRG2wU5USziF3OC3YYZIXiPe41IltP7LSUmyRO/F6jAcUj
ZmRP2sxhIjY/77nQbx1F3ZMF2i91CRyaIfyd2pC8pwA4VElBTZaP9j3xXEsA8VIX
F/PSVcDsm3GoxVkwQbJTr54GedsRMoex574rvt8iujiNQ7Cb0uXWFIfnlD1thnne
4ws5ekuVhT6lq1KDB2z4e/pN2cOEzzSmfJJK1AWS79R4sAO8Fm/8cpWx6MRhlAbv
HwIDAQAB
-----END PUBLIC KEY-----`

	fingerprint2 = "SHA256:EM4Jt7RjoQIPqpRFTadBCQkdzu+G4tq1RWd3f+I6nRg"

	validKey3 = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA06q+yHMtXDj3qa3qELcg
bS/48HWbylEi+smx+xa8yupMTMtne6WFvxiS3lU/+TXQj+hdHzwpLj+W24QCON1o
JqxYDLWVJ2YpmrwkU/IDbhoPKfpYchy6Zmg2bnr93FDcvc4oL2/UYaiG+3w8fS+D
BcHug7ILLmY5RnwqzdcYfQ5waX2QCK75kmtB+TBqtS3xAr2m2omdla91YeARSu3O
lVjB6h9QNfbR6KCZRalMWlNGpp0tG0faU9mEescY4zfqt2inQFAr+MuXjJhg0tW8
kO6LskiW1+SbBlNrJeQDXUjC/vz6/8X1DvDeczd9tqbAxfV57yRjIxkfsDYxehai
6QIDAQAB
-----END PUBLIC KEY-----`

	fingerprint3 = "SHA256:DGsuFb8nHgtg88dwIsTnGL3J8Hx+yCksl0WEBCbm5Zc"

	invalidKey = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtzMnDEQPd75QZByogNlB
NY2auyr4sy8UNTDARs79Edq/Jw5tb7ub412mOB61mVrcuFZW6xfmCRt0ILgoaT66
Tp1RpuEfghD+e7bYZ+Q2pckC1ZaVPIVVf/ZcCZ0tKQHoD8EpyyFINKjCh516VrCx
KuOm2fALPB/xDwDBEdeVJlh5/3HHP2V35scdvDRkvr2qkcvhzoy0+7wUWFRZ2n6H
TFrxMHQoHg0tutAJEkjsMw9xfN7V07c952SHNRZvu80V5EEpnKw/iYKXUjCmoXm8
tpJv5kXH6XPgfvOirSbTfuo+0VGqVIx9gcomzJ0I5WfGTD22dAxDiRT7q7KZnNgt
TwIDAQAB
`
)

func compareUnorderedFingerprints(tb testing.TB, got, want []string) {
	tb.Helper()

	tb.Logf("compareUnorderedFingerprints: got %v, want %v", got, want)

	if len(got) != len(want) {
		tb.Errorf("len mismatch: got %d, want %d", len(got), len(want))
	}

	for _, s := range got {
		var found bool
		for _, t := range want {
			if s == t {
				found = true
				break
			}
		}
		if !found {
			tb.Errorf("%q in got not found in want", s)
		}
	}

	for _, t := range want {
		var found bool
		for _, s := range got {
			if s == t {
				found = true
				break
			}
		}
		if !found {
			tb.Errorf("%q in want not found in got", t)
		}
	}
}

func TestParseVersionedKeys(t *testing.T) {
	for _, c := range []struct {
		label            string
		secret           secrets.VersionedSecret
		nopLogger        bool
		expectNil        bool
		firstFingerprint string
		fingerprints     []string
	}{
		{
			label: "all-valid",
			secret: secrets.VersionedSecret{
				Current:  []byte(validKey1),
				Previous: []byte(validKey2),
				Next:     []byte(validKey3),
			},
			firstFingerprint: fingerprint1,
			fingerprints: []string{
				fingerprint1,
				fingerprint2,
				fingerprint3,
			},
		},
		{
			label: "invalid-current",
			secret: secrets.VersionedSecret{
				Current:  []byte(invalidKey),
				Previous: []byte(validKey2),
				Next:     []byte(validKey3),
			},
			nopLogger:        true,
			firstFingerprint: fingerprint2,
			fingerprints: []string{
				fingerprint2,
				fingerprint3,
			},
		},
		{
			label: "only-current",
			secret: secrets.VersionedSecret{
				Current: []byte(validKey1),
			},
			firstFingerprint: fingerprint1,
			fingerprints: []string{
				fingerprint1,
			},
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			var logger log.Wrapper
			if c.nopLogger {
				logger = log.NopWrapper
			} else {
				logger = log.TestWrapper(t)
			}
			keys := parseVersionedKeys(context.Background(), c.secret, logger)
			if c.expectNil {
				if keys != nil {
					t.Errorf("Expected nil result, got %v", keys)
					return
				}
			} else {
				if keys == nil {
					t.Error("Unexpected nil result")
					return
				}
			}
			fingerprints := make([]string, 0, len(keys.m))
			for k := range keys.m {
				fingerprints = append(fingerprints, k)
			}
			compareUnorderedFingerprints(t, fingerprints, c.fingerprints)

			fingerprint, err := RSAPublicKeyFingerprint(keys.first)
			if err != nil {
				t.Errorf("Unable to calculate fingerprint of keys.first: %v", err)
			}
			if fingerprint != c.firstFingerprint {
				t.Errorf("keys.first fingerprint got %q, want %q", fingerprint, c.firstFingerprint)
			}
		})
	}
}
