package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

const maxSavedSearchNameLen = 80

var (
	ErrSavedSearchNotFound   = errors.New("saved search not found")
	ErrSavedSearchNameEmpty  = errors.New("saved search name required")
	ErrSavedSearchNameTaken  = errors.New("saved search name taken")
	ErrSavedSearchInvalid    = errors.New("invalid saved search")
	ErrListSavedSearches     = errors.New("list saved searches")
	ErrCreateSavedSearch     = errors.New("create saved search")
	ErrUpdateSavedSearch     = errors.New("update saved search")
	ErrDeleteSavedSearch     = errors.New("delete saved search")
)

type SavedSearchInput struct {
	Name      string
	Q         string
	Type      string
	Location  string
	Recency   string
	SavedOnly bool
	SortBy    string
	SortDir   string
}

type SavedSearchStore interface {
	ListUserSavedSearches(ctx context.Context, userID pgtype.UUID) ([]db.UserSavedSearch, error)
	GetUserSavedSearch(ctx context.Context, arg db.GetUserSavedSearchParams) (db.UserSavedSearch, error)
	CreateUserSavedSearch(ctx context.Context, arg db.CreateUserSavedSearchParams) (db.UserSavedSearch, error)
	UpdateUserSavedSearch(ctx context.Context, arg db.UpdateUserSavedSearchParams) (db.UserSavedSearch, error)
	DeleteUserSavedSearch(ctx context.Context, arg db.DeleteUserSavedSearchParams) (int64, error)
}

type SavedSearchService struct {
	store SavedSearchStore
}

func NewSavedSearchService(store SavedSearchStore) *SavedSearchService {
	return &SavedSearchService{store: store}
}

func (s *SavedSearchService) List(ctx context.Context, userID pgtype.UUID) ([]db.UserSavedSearch, error) {
	searches, err := s.store.ListUserSavedSearches(ctx, userID)
	if err != nil {
		return nil, errors.Join(ErrListSavedSearches, err)
	}
	if searches == nil {
		return []db.UserSavedSearch{}, nil
	}
	return searches, nil
}

func (s *SavedSearchService) Create(ctx context.Context, userID pgtype.UUID, input SavedSearchInput) (db.UserSavedSearch, error) {
	params, err := toSavedSearchParams(userID, input)
	if err != nil {
		return db.UserSavedSearch{}, err
	}

	created, err := s.store.CreateUserSavedSearch(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return db.UserSavedSearch{}, ErrSavedSearchNameTaken
		}
		return db.UserSavedSearch{}, errors.Join(ErrCreateSavedSearch, err)
	}
	return created, nil
}

func (s *SavedSearchService) Update(ctx context.Context, userID, searchID pgtype.UUID, input SavedSearchInput) (db.UserSavedSearch, error) {
	params, err := toSavedSearchUpdateParams(userID, searchID, input)
	if err != nil {
		return db.UserSavedSearch{}, err
	}

	updated, err := s.store.UpdateUserSavedSearch(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.UserSavedSearch{}, ErrSavedSearchNotFound
		}
		if isUniqueViolation(err) {
			return db.UserSavedSearch{}, ErrSavedSearchNameTaken
		}
		return db.UserSavedSearch{}, errors.Join(ErrUpdateSavedSearch, err)
	}
	return updated, nil
}

func (s *SavedSearchService) Delete(ctx context.Context, userID, searchID pgtype.UUID) error {
	rows, err := s.store.DeleteUserSavedSearch(ctx, db.DeleteUserSavedSearchParams{
		ID:     searchID,
		UserID: userID,
	})
	if err != nil {
		return errors.Join(ErrDeleteSavedSearch, err)
	}
	if rows == 0 {
		return ErrSavedSearchNotFound
	}
	return nil
}

func toSavedSearchParams(userID pgtype.UUID, input SavedSearchInput) (db.CreateUserSavedSearchParams, error) {
	normalized, err := normalizeSavedSearchInput(input)
	if err != nil {
		return db.CreateUserSavedSearchParams{}, err
	}
	return db.CreateUserSavedSearchParams{
		UserID:    userID,
		Name:      normalized.Name,
		Q:         normalized.Q,
		JobType:   normalized.Type,
		Location:  normalized.Location,
		Recency:   normalized.Recency,
		SavedOnly: normalized.SavedOnly,
		SortBy:    normalized.SortBy,
		SortDir:   normalized.SortDir,
	}, nil
}

func toSavedSearchUpdateParams(userID, searchID pgtype.UUID, input SavedSearchInput) (db.UpdateUserSavedSearchParams, error) {
	normalized, err := normalizeSavedSearchInput(input)
	if err != nil {
		return db.UpdateUserSavedSearchParams{}, err
	}
	return db.UpdateUserSavedSearchParams{
		ID:        searchID,
		UserID:    userID,
		Name:      normalized.Name,
		Q:         normalized.Q,
		JobType:   normalized.Type,
		Location:  normalized.Location,
		Recency:   normalized.Recency,
		SavedOnly: normalized.SavedOnly,
		SortBy:    normalized.SortBy,
		SortDir:   normalized.SortDir,
	}, nil
}

func normalizeSavedSearchInput(input SavedSearchInput) (SavedSearchInput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return SavedSearchInput{}, ErrSavedSearchNameEmpty
	}
	if len(name) > maxSavedSearchNameLen {
		return SavedSearchInput{}, ErrSavedSearchInvalid
	}

	recency := strings.ToLower(strings.TrimSpace(input.Recency))
	switch recency {
	case "", "24h", "3d", "7d":
	default:
		return SavedSearchInput{}, ErrSavedSearchInvalid
	}

	sortBy, sortDir, err := ParseSort(input.SortBy, input.SortDir)
	if err != nil {
		return SavedSearchInput{}, ErrSavedSearchInvalid
	}

	return SavedSearchInput{
		Name:      name,
		Q:         strings.TrimSpace(input.Q),
		Type:      strings.TrimSpace(input.Type),
		Location:  strings.TrimSpace(input.Location),
		Recency:   recency,
		SavedOnly: input.SavedOnly,
		SortBy:    sortBy,
		SortDir:   sortDir,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
