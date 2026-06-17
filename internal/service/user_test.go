package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type mockUserStore struct {
	count       int64
	countErr    error
	createErr   error
	createUser  func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	createCalls []db.CreateUserParams
}

func (m *mockUserStore) GetUserCount(ctx context.Context) (int64, error) {
	return m.count, m.countErr
}

func (m *mockUserStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	m.createCalls = append(m.createCalls, arg)
	if m.createUser != nil {
		return m.createUser(ctx, arg)
	}
	if m.createErr != nil {
		return db.User{}, m.createErr
	}
	return db.User{}, nil
}

func TestUserService_Register(t *testing.T) {
	t.Run("rejects missing fields", func(t *testing.T) {
		store := &mockUserStore{}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{Email: "a@b.com"})
		if !errors.Is(err, ErrMissingFields) {
			t.Fatalf("err = %v, want ErrMissingFields", err)
		}
		if len(store.createCalls) != 0 {
			t.Fatal("expected store not to be called")
		}
	})

	t.Run("bootstrap privileges for first user", func(t *testing.T) {
		store := &mockUserStore{count: 0}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  "secure-password",
		})
		if err != nil {
			t.Fatal(err)
		}

		created := store.createCalls[0]
		if !boolVal(created.IsAdmin) || !boolVal(created.IsActive) || !boolVal(created.IsPremium) {
			t.Fatal("expected bootstrap privileges")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(created.Password), []byte("secure-password")); err != nil {
			t.Fatal("expected bcrypt hash")
		}
	})

	t.Run("default privileges for subsequent user", func(t *testing.T) {
		store := &mockUserStore{count: 2}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Grace",
			LastName:  "Hopper",
			Email:     "grace@example.com",
			Password:  "secure-password",
		})
		if err != nil {
			t.Fatal(err)
		}

		created := store.createCalls[0]
		if boolVal(created.IsAdmin) || boolVal(created.IsActive) || boolVal(created.IsPremium) {
			t.Fatal("expected default privileges")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		store := &mockUserStore{
			createUser: func(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
				return db.User{}, &pgconn.PgError{Code: "23505"}
			},
		}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "dup@example.com",
			Password:  "secure-password",
		})
		if !errors.Is(err, ErrUserExists) {
			t.Fatalf("err = %v, want ErrUserExists", err)
		}
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		store := &mockUserStore{}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "not-an-email",
			Password:  "secure-password",
		})
		if !errors.Is(err, ErrInvalidEmail) {
			t.Fatalf("err = %v, want ErrInvalidEmail", err)
		}
	})

	t.Run("rejects weak password", func(t *testing.T) {
		store := &mockUserStore{}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@example.com",
			Password:  "short",
		})
		if !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("err = %v, want ErrWeakPassword", err)
		}
	})

	t.Run("normalizes email", func(t *testing.T) {
		store := &mockUserStore{count: 1}
		svc := NewUserService(store, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "  Ada@Example.COM ",
			Password:  "secure-password",
		})
		if err != nil {
			t.Fatal(err)
		}
		if store.createCalls[0].Email != "ada@example.com" {
			t.Fatalf("email = %q, want ada@example.com", store.createCalls[0].Email)
		}
	})

	t.Run("count users error", func(t *testing.T) {
		svc := NewUserService(&mockUserStore{countErr: errors.New("db down")}, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "a@b.com",
			Password:  "password1",
		})
		if !errors.Is(err, ErrCountUsers) {
			t.Fatalf("err = %v, want ErrCountUsers", err)
		}
	})

	t.Run("hash password error", func(t *testing.T) {
		svc := NewUserService(&mockUserStore{}, func(string) (string, error) {
			return "", errors.New("hash failed")
		})

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "a@b.com",
			Password:  "password1",
		})
		if !errors.Is(err, ErrHashPassword) {
			t.Fatalf("err = %v, want ErrHashPassword", err)
		}
	})

	t.Run("create user error", func(t *testing.T) {
		svc := NewUserService(&mockUserStore{createErr: errors.New("insert failed")}, nil)

		err := svc.Register(context.Background(), RegisterInput{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "a@b.com",
			Password:  "password1",
		})
		if !errors.Is(err, ErrCreateUser) {
			t.Fatalf("err = %v, want ErrCreateUser", err)
		}
	})
}

func boolVal(v *bool) bool {
	return v != nil && *v
}
