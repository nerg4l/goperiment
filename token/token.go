package token

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"hash"
	"net/http"
	"strings"
)

func randomString(l int) (string, error) {
	b := make([]byte, l)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// generally base64 encoding makes the len 1.3 times longer
	str := base64.RawURLEncoding.EncodeToString(b)
	return str[:l], nil
}

// Generate creates a new token with 40 char length.
func Generate() (string, error) {
	return randomString(40)
}

// GenerateHashed creates a new hashed token.
func GenerateHashed(h hash.Hash) (string, error) {
	s, err := Generate()
	if err != nil {
		return "", err
	}
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ErrNoBearerToken is returned by GetRequestBearer when no authorization header
// is set or the authorization header is not a bearer token.
var ErrNoBearerToken = errors.New("token: no bearer token in request header")

// GetRequestBearer returns the bearer token from the request when available.
// It returns an ErrNoBearerToken error when missing or a hex error when the
// token is not a valid hex.
func GetRequestBearer(r *http.Request) ([]byte, error) {
	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		return nil, ErrNoBearerToken
	}
	return hex.DecodeString(strings.TrimPrefix(authorization, "Bearer "))
}
