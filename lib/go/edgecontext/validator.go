package edgecontext

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/reddit/baseplate.go/log"
	"github.com/reddit/baseplate.go/secrets"
	"golang.org/x/crypto/ssh"
)

type keysType struct {
	// map of kid -> pub key.
	m map[string]*rsa.PublicKey

	// when either kid header does not exist in the jwt token,
	// or the kid is not present in the map,
	// we fallback to the first (usually current) key.
	first *rsa.PublicKey
}

func (kt *keysType) getKey(kid string) *rsa.PublicKey {
	if key := kt.m[kid]; key != nil {
		return key
	}
	return kt.first
}

const (
	authenticationPubKeySecretPath = "secret/authentication/public-key"
	jwtAlg                         = "RS256"
)

// JWTHeaderKeyID is the JWT header for the key id,
// as defined in RFC 7517 section 4.5.
const JWTHeaderKeyID = "kid"

// ErrNoPublicKeysLoaded is an error returned by ValidateToken indicates that
// the function is called before any public keys are loaded from secrets.
var ErrNoPublicKeysLoaded = errors.New("edgecontext.ValidateToken: no public keys loaded")

// ErrEmptyToken is an error returned by ValidateToken indicates that the JWT
// token is empty string.
var ErrEmptyToken = errors.New("edgecontext.ValidateToken: empty JWT token")

// ValidateToken parses and validates a jwt token, and return the decoded
// AuthenticationToken.
func (impl *Impl) ValidateToken(token string) (*AuthenticationToken, error) {
	keys, ok := impl.keysValue.Load().(*keysType)
	if !ok {
		// This would only happen when all previous middleware parsing failed.
		return nil, ErrNoPublicKeysLoaded
	}

	if token == "" {
		// If we don't do the special handling here,
		// jwt.ParseWithClaims below will return an error with message
		// "token contains an invalid number of segments".
		// Also that's still true, it's less obvious what's actually going on.
		// Returning different error for empty token can also help highlighting
		// other invalid tokens that actually causes that invalid number of segments
		// error.
		return nil, ErrEmptyToken
	}

	tok, err := jwt.ParseWithClaims(
		token,
		&AuthenticationToken{},
		func(jt *jwt.Token) (interface{}, error) {
			kid, _ := jt.Header[JWTHeaderKeyID].(string)
			return keys.getKey(kid), nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !tok.Valid {
		return nil, jwt.NewValidationError("invalid token", 0)
	}

	if tok.Method.Alg() != jwtAlg {
		return nil, jwt.NewValidationError("wrong signing method", 0)
	}

	if claims, ok := tok.Claims.(*AuthenticationToken); ok {
		return claims, nil
	}

	return nil, jwt.NewValidationError("invalid token type", 0)
}

func (impl *Impl) validatorMiddleware(next secrets.SecretHandlerFunc) secrets.SecretHandlerFunc {
	return func(sec *secrets.Secrets) {
		defer next(sec)

		versioned, err := sec.GetVersionedSecret(authenticationPubKeySecretPath)
		if err != nil {
			impl.logger.Log(context.Background(), fmt.Sprintf(
				"Failed to get secrets %q: %v",
				authenticationPubKeySecretPath,
				err,
			))
			return
		}

		keys := parseVersionedKeys(context.Background(), versioned, impl.logger)
		if keys != nil {
			impl.keysValue.Store(keys)
		}
	}
}

func parseVersionedKeys(ctx context.Context, versioned secrets.VersionedSecret, logger log.Wrapper) *keysType {
	all := versioned.GetAll()
	keys := &keysType{
		m: make(map[string]*rsa.PublicKey, len(all)),
	}
	for i, v := range all {
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(v))
		if err != nil {
			logger.Log(ctx, fmt.Sprintf(
				"Failed to parse key #%d: %v",
				i,
				err,
			))
		} else {
			if keys.first == nil {
				keys.first = key
			}
			if fingerprint, err := RSAPublicKeyFingerprint(key); err != nil {
				logger.Log(ctx, fmt.Sprintf(
					"Failed to get fingerprint of key #%d: %v",
					i,
					err,
				))
			} else {
				keys.m[fingerprint] = key
			}
		}
	}
	if keys.first == nil {
		logger.Log(ctx, "No valid keys in secrets store.")
		return nil
	}
	return keys
}

// RSAPublicKeyFingerprint calculates the fingerprint of an RSA public key,
// using ssh.FingerprintSHA256:
// https://pkg.go.dev/golang.org/x/crypto/ssh#FingerprintSHA256
func RSAPublicKeyFingerprint(pubKey *rsa.PublicKey) (string, error) {
	key, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(key), nil
}
