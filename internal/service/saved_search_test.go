package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

type mockSavedSearchStore struct {
	searches []db.UserSavedSearch
	listErr  error
	create   func(ctx context.Context, arg db.CreateUserSavedSearchParams) (db.UserSavedSearch, error)
	update   func(ctx context.Context, arg db.UpdateUserSavedSearchParams) (db.UserSavedSearch, error)
	delete   func(ctx context.Context, arg db.DeleteUserSavedSearchParams) (int64, error)
}

func (m *mockSavedSearchStore) ListUserSavedSearches(ctx context.Context, userID pgtype.UUID) ([]db.UserSavedSearch, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.searches, nil
}

func (m *mockSavedSearchStore) GetUserSavedSearch(ctx context.Context, arg db.GetUserSavedSearchParams) (db.UserSavedSearch, error) {
	for _, search := range m.searches {
		if search.ID == arg.ID && search.UserID == arg.UserID {
			return search, nil
		}
	}
	return db.UserSavedSearch{}, ErrSavedSearchNotFound
}

func (m *mockSavedSearchStore) CreateUserSavedSearch(ctx context.Context, arg db.CreateUserSavedSearchParams) (db.UserSavedSearch, error) {
	if m.create != nil {
		return m.create(ctx, arg)
	}
	var id pgtype.UUID
	_ = id.Scan("11111111-1111-1111-1111-111111111111")
	return db.UserSavedSearch{
		ID:        id,
		UserID:    arg.UserID,
		Name:      arg.Name,
		Q:         arg.Q,
		JobType:   arg.JobType,
		Location:  arg.Location,
		Recency:   arg.Recency,
		SavedOnly: arg.SavedOnly,
		SortBy:    arg.SortBy,
		SortDir:   arg.SortDir,
	}, nil
}

func (m *mockSavedSearchStore) UpdateUserSavedSearch(ctx context.Context, arg db.UpdateUserSavedSearchParams) (db.UserSavedSearch, error) {
	if m.update != nil {
		return m.update(ctx, arg)
	}
	return db.UserSavedSearch{
		ID:        arg.ID,
		UserID:    arg.UserID,
		Name:      arg.Name,
		Q:         arg.Q,
		JobType:   arg.JobType,
		Location:  arg.Location,
		Recency:   arg.Recency,
		SavedOnly: arg.SavedOnly,
		SortBy:    arg.SortBy,
		SortDir:   arg.SortDir,
	}, nil
}

func (m *mockSavedSearchStore) DeleteUserSavedSearch(ctx context.Context, arg db.DeleteUserSavedSearchParams) (int64, error) {
	if m.delete != nil {
		return m.delete(ctx, arg)
	}
	return 1, nil
}

func TestSavedSearchService_Create(t *testing.T) {
	var userID pgtype.UUID
	_ = userID.Scan("22222222-2222-2222-2222-222222222222")

	svc := NewSavedSearchService(&mockSavedSearchStore{})
	created, err := svc.Create(context.Background(), userID, SavedSearchInput{
		Name:     "NYC interns",
		Q:        "intern",
		Location: "New York, NY",
		Type:     "Internship",
		SortBy:   "posted",
		SortDir:  "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "NYC interns" || created.Q != "intern" {
		t.Fatalf("created = %+v", created)
	}

	_, err = svc.Create(context.Background(), userID, SavedSearchInput{})
	if err != ErrSavedSearchNameEmpty {
		t.Fatalf("err = %v", err)
	}
}

func TestSavedSearchService_Create_nameTaken(t *testing.T) {
	var userID pgtype.UUID
	_ = userID.Scan("22222222-2222-2222-2222-222222222222")

	svc := NewSavedSearchService(&mockSavedSearchStore{
		create: func(ctx context.Context, arg db.CreateUserSavedSearchParams) (db.UserSavedSearch, error) {
			return db.UserSavedSearch{}, &pgconn.PgError{Code: "23505"}
		},
	})
	_, err := svc.Create(context.Background(), userID, SavedSearchInput{Name: "Taken"})
	if err != ErrSavedSearchNameTaken {
		t.Fatalf("err = %v", err)
	}
}

func TestSavedSearchService_Delete_notFound(t *testing.T) {
	var userID, searchID pgtype.UUID
	_ = userID.Scan("22222222-2222-2222-2222-222222222222")
	_ = searchID.Scan("33333333-3333-3333-3333-333333333333")

	svc := NewSavedSearchService(&mockSavedSearchStore{
		delete: func(ctx context.Context, arg db.DeleteUserSavedSearchParams) (int64, error) {
			return 0, nil
		},
	})
	err := svc.Delete(context.Background(), userID, searchID)
	if err != ErrSavedSearchNotFound {
		t.Fatalf("err = %v", err)
	}
}
