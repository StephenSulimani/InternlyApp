package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

func TestTokenIssuer_IssueAndParse(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Hour)

	var id pgtype.UUID
	if err := id.Scan("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Fatal(err)
	}

	user := db.User{
		ID:        id,
		Email:     "ada@example.com",
		FirstName: "Ada",
		LastName:  "Lovelace",
		IsActive:  true,
	}

	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != id.String() {
		t.Fatalf("user_id = %q", claims.UserID)
	}
	if claims.Email != user.Email {
		t.Fatalf("email = %q", claims.Email)
	}
}

func TestTokenIssuer_ParseRejectsInvalidToken(t *testing.T) {
	issuer := NewTokenIssuer("test-secret", time.Hour)
	_, err := issuer.Parse("not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}
