package auth

import "testing"

func TestHashPasswordAndCheck(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPassword("correct-horse-battery-staple", hash) {
		t.Fatal("expected password to match hash")
	}

	if CheckPassword("wrong-password", hash) {
		t.Fatal("expected password mismatch")
	}
}
