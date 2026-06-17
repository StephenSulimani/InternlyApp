package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
)

const MinPasswordLength = 8

var (
	ErrMissingFields = errors.New("missing required fields")
	ErrInvalidEmail  = errors.New("invalid email")
	ErrWeakPassword  = errors.New("weak password")
	ErrUserExists    = errors.New("user already exists")
	ErrCountUsers    = errors.New("count users")
	ErrHashPassword  = errors.New("hash password")
	ErrCreateUser    = errors.New("create user")
)

type PasswordHasher func(password string) (string, error)

type UserService struct {
	store  UserStore
	hasher PasswordHasher
}

func NewUserService(store UserStore, hasher PasswordHasher) *UserService {
	if hasher == nil {
		hasher = auth.HashPassword
	}
	return &UserService{store: store, hasher: hasher}
}

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) error {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = normalizeEmail(input.Email)

	if input.FirstName == "" || input.LastName == "" || input.Email == "" || input.Password == "" {
		return ErrMissingFields
	}
	if !isValidEmail(input.Email) {
		return ErrInvalidEmail
	}
	if len(input.Password) < MinPasswordLength {
		return ErrWeakPassword
	}

	count, err := s.store.GetUserCount(ctx)
	if err != nil {
		return errors.Join(ErrCountUsers, err)
	}

	hashedPassword, err := s.hasher(input.Password)
	if err != nil {
		return errors.Join(ErrHashPassword, err)
	}

	isActive, isAdmin, isPremium := false, false, false
	if count == 0 {
		isActive, isAdmin, isPremium = true, true, true
	}

	_, err = s.store.CreateUser(ctx, db.CreateUserParams{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  hashedPassword,
		IsActive:  &isActive,
		IsAdmin:   &isAdmin,
		IsPremium: &isPremium,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return ErrUserExists
		}
		return errors.Join(ErrCreateUser, err)
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email
}
