package token

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"hash"
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
