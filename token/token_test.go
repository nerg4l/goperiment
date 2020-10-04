package token

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"net/http/httptest"
	"reflect"
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

func TestGetRequestBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if _, err := GetRequestBearer(req); err != ErrNoBearerToken {
		t.Errorf("GetRequestBearer() error = %v, wantErr %v", err, ErrNoBearerToken)
	}
	req.Header.Set("Authorization", "any")
	if _, err := GetRequestBearer(req); err != ErrNoBearerToken {
		t.Errorf("GetRequestBearer() error = %v, wantErr %v", err, ErrNoBearerToken)
	}
	req.Header.Set("Authorization", "Bearer 0g")
	if _, err := GetRequestBearer(req); err != hex.InvalidByteError('g') {
		t.Errorf("GetRequestBearer() error = %v, wantErr %v", err, hex.InvalidByteError('g'))
	}
	req.Header.Set("Authorization", "Bearer 0001020304050607")
	got, err := GetRequestBearer(req)
	if err != nil {
		t.Errorf("GetRequestBearer() error = %v, wantErr %v", err, nil)
	}
	want := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetRequestBearer() got = %v, want %v", got, want)
	}
}
