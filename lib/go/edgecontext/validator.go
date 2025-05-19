package edgecontext

import (
	"crypto/rsa"

	v0 "github.com/reddit/edgecontext/lib/go/internal/v0"
)

// RSAPublicKeyFingerprint calculates the fingerprint of an RSA public key,
// using ssh.FingerprintSHA256:
// https://pkg.go.dev/golang.org/x/crypto/ssh#FingerprintSHA256
func RSAPublicKeyFingerprint(pubKey *rsa.PublicKey) (string, error) {
	return v0.RSAPublicKeyFingerprint(pubKey)
}
