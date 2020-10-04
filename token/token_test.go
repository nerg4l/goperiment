package token

import (
	"crypto/sha256"
	"crypto/sha512"
	"testing"
)

func TestGenerate(t *testing.T) {
	token, _ := Generate()
	wantLen := 40
	if len(token) != wantLen {
		t.Fatalf("Generate() = %v, wantLen %v", token, wantLen)
	}
}

func TestGenerateHashed(t *testing.T) {
	token, _ := GenerateHashed(sha256.New())
	wantLen := 64
	if len(token) != wantLen {
		t.Fatalf("Generate() = %v, wantLen %v", token, wantLen)
	}
	token, _ = GenerateHashed(sha512.New())
	wantLen = 128
	if len(token) != wantLen {
		t.Fatalf("Generate() = %v, wantLen %v", token, wantLen)
	}
}
