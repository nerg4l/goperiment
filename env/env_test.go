package env

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	const smallpox = "SMALLPOX"
	value := Get(smallpox, "vaccine")
	if value != "vaccine" {
		t.Fatalf("Get() = %v, want %v", value, "vaccine")
	}
	defer os.Unsetenv(smallpox)
	err := os.Setenv(smallpox, "vaccine")
	if err != nil {
		t.Fatalf("failed to release smallpox vaccine")
	}
	value = Get(smallpox, "virus")
	if value != "vaccine" {
		t.Fatalf("Get() = %v, want %v", value, "vaccine")
	}
}
