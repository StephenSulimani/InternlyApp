package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrMissingFields = errors.New("missing required fields")
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
		hasher = hashPassword
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
	if input.FirstName == "" || input.LastName == "" || input.Email == "" || input.Password == "" {
		return ErrMissingFields
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

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}
