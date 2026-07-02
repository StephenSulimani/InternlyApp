package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
)

type mockUserStore struct {
	count          int64
	countErr       error
	createErr      error
	createUser     func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	getUserByEmail func(ctx context.Context, email string) (db.User, error)
	createCalls    []db.CreateUserParams
}

func (m *mockUserStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.getUserByEmail != nil {
		return m.getUserByEmail(ctx, email)
	}
	return db.User{}, pgx.ErrNoRows
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
		if !created.IsAdmin || !created.IsActive || !created.IsPremium {
			t.Fatal("expected bootstrap privileges")
		}
		if !auth.CheckPassword("secure-password", created.Password) {
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
		if created.IsAdmin || created.IsActive || created.IsPremium {
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

func TestUserService_SeedAdmin(t *testing.T) {
	t.Run("creates active admin regardless of user count", func(t *testing.T) {
		store := &mockUserStore{count: 5}
		svc := NewUserService(store, nil)

		_, err := svc.SeedAdmin(context.Background(), RegisterInput{
			FirstName: "Stephen",
			LastName:  "Sulimani",
			Email:     "stephen@example.com",
			Password:  "password1",
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(store.createCalls) != 1 {
			t.Fatalf("createCalls = %d, want 1", len(store.createCalls))
		}
		if !store.createCalls[0].IsAdmin || !store.createCalls[0].IsActive || !store.createCalls[0].IsPremium {
			t.Fatal("expected admin flags on create params")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		store := &mockUserStore{
			createErr: &pgconn.PgError{Code: "23505"},
		}
		svc := NewUserService(store, nil)

		_, err := svc.SeedAdmin(context.Background(), RegisterInput{
			FirstName: "Stephen",
			LastName:  "Sulimani",
			Email:     "stephen@example.com",
			Password:  "password1",
		})
		if !errors.Is(err, ErrUserExists) {
			t.Fatalf("err = %v, want ErrUserExists", err)
		}
	})
}

func TestUserService_Login(t *testing.T) {
	t.Run("returns user for valid credentials", func(t *testing.T) {
		hash, err := auth.HashPassword("secure-password")
		if err != nil {
			t.Fatal(err)
		}
		store := &mockUserStore{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{
					Email:    email,
					Password: hash,
					IsActive: true,
				}, nil
			},
		}
		svc := NewUserService(store, nil)

		user, err := svc.Login(context.Background(), LoginInput{
			Email:    "Ada@Example.COM",
			Password: "secure-password",
		})
		if err != nil {
			t.Fatal(err)
		}
		if user.Email != "ada@example.com" {
			t.Fatalf("email = %q", user.Email)
		}
	})

	t.Run("rejects missing fields", func(t *testing.T) {
		svc := NewUserService(&mockUserStore{}, nil)
		_, err := svc.Login(context.Background(), LoginInput{Email: "a@b.com"})
		if !errors.Is(err, ErrMissingFields) {
			t.Fatalf("err = %v, want ErrMissingFields", err)
		}
	})

	t.Run("rejects unknown email", func(t *testing.T) {
		store := &mockUserStore{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{}, pgx.ErrNoRows
			},
		}
		svc := NewUserService(store, nil)
		_, err := svc.Login(context.Background(), LoginInput{
			Email:    "missing@example.com",
			Password: "secure-password",
		})
		if !errors.Is(err, ErrInvalidEmailOrPassword) {
			t.Fatalf("err = %v, want ErrInvalidEmailOrPassword", err)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		hash, err := auth.HashPassword("secure-password")
		if err != nil {
			t.Fatal(err)
		}
		store := &mockUserStore{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{Email: email, Password: hash, IsActive: true}, nil
			},
		}
		svc := NewUserService(store, nil)
		_, err = svc.Login(context.Background(), LoginInput{
			Email:    "ada@example.com",
			Password: "wrong-password",
		})
		if !errors.Is(err, ErrInvalidEmailOrPassword) {
			t.Fatalf("err = %v, want ErrInvalidEmailOrPassword", err)
		}
	})

	t.Run("rejects inactive user", func(t *testing.T) {
		hash, err := auth.HashPassword("secure-password")
		if err != nil {
			t.Fatal(err)
		}
		store := &mockUserStore{
			getUserByEmail: func(ctx context.Context, email string) (db.User, error) {
				return db.User{Email: email, Password: hash, IsActive: false}, nil
			},
		}
		svc := NewUserService(store, nil)
		_, err = svc.Login(context.Background(), LoginInput{
			Email:    "ada@example.com",
			Password: "secure-password",
		})
		if !errors.Is(err, ErrUserInactive) {
			t.Fatalf("err = %v, want ErrUserInactive", err)
		}
	})
}
