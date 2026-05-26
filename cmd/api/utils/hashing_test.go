package utils

import "testing"

func TestHashPasswordAndCheck(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPasswordHash("correct-horse-battery-staple", hash) {
		t.Fatal("expected password to match hash")
	}

	if CheckPasswordHash("wrong-password", hash) {
		t.Fatal("expected password mismatch")
	}
}
