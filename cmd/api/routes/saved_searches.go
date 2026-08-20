package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

type savedSearchBody struct {
	Name     string `json:"name"`
	Q        string `json:"q"`
	Type     string `json:"type"`
	Location string `json:"location"`
	Recency  string `json:"recency"`
	Saved    bool   `json:"saved"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

func (b savedSearchBody) toInput() service.SavedSearchInput {
	return service.SavedSearchInput{
		Name:      b.Name,
		Q:         b.Q,
		Type:      b.Type,
		Location:  b.Location,
		Recency:   b.Recency,
		SavedOnly: b.Saved,
		SortBy:    b.Sort,
		SortDir:   b.Order,
	}
}

func ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	svc, user, log, ok := savedSearchDeps(w, r)
	if !ok {
		return
	}

	searches, err := svc.List(r.Context(), user.ID)
	if err != nil {
		writeSavedSearchError(w, log, err)
		return
	}

	types.WriteSavedSearches(w, types.SavedSearchesFrom(searches))
}

func CreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	svc, user, log, ok := savedSearchDeps(w, r)
	if !ok {
		return
	}

	body, ok := savedSearchBodyFromRequest(w, r)
	if !ok {
		return
	}

	created, err := svc.Create(r.Context(), user.ID, body.toInput())
	if err != nil {
		writeSavedSearchError(w, log, err)
		return
	}

	types.WriteSavedSearch(w, http.StatusCreated, "Saved search created", types.SavedSearchFrom(created))
}

func UpdateSavedSearch(w http.ResponseWriter, r *http.Request) {
	svc, user, log, ok := savedSearchDeps(w, r)
	if !ok {
		return
	}

	searchID, err := savedSearchIDFromPath(r)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, "Invalid saved search id")
		return
	}

	body, ok := savedSearchBodyFromRequest(w, r)
	if !ok {
		return
	}

	updated, err := svc.Update(r.Context(), user.ID, searchID, body.toInput())
	if err != nil {
		writeSavedSearchError(w, log, err)
		return
	}

	types.WriteSavedSearch(w, http.StatusOK, "Saved search updated", types.SavedSearchFrom(updated))
}

func DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	svc, user, log, ok := savedSearchDeps(w, r)
	if !ok {
		return
	}

	searchID, err := savedSearchIDFromPath(r)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, "Invalid saved search id")
		return
	}

	if err := svc.Delete(r.Context(), user.ID, searchID); err != nil {
		writeSavedSearchError(w, log, err)
		return
	}

	types.WriteSuccess(w, http.StatusOK, "Saved search deleted")
}

func savedSearchDeps(w http.ResponseWriter, r *http.Request) (*service.SavedSearchService, db.User, *zap.SugaredLogger, bool) {
	log, ok := middleware.LoggerFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return nil, db.User{}, nil, false
	}

	svc, ok := savedSearchServiceFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return nil, db.User{}, nil, false
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return nil, db.User{}, nil, false
	}

	return svc, user, log, true
}

func savedSearchBodyFromRequest(w http.ResponseWriter, r *http.Request) (savedSearchBody, bool) {
	body, ok := middleware.BodyFromContext(r.Context())
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return savedSearchBody{}, false
	}

	var parsed savedSearchBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		types.WriteError(w, http.StatusBadRequest, "Error parsing request body")
		return savedSearchBody{}, false
	}
	return parsed, true
}

func savedSearchIDFromPath(r *http.Request) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(mux.Vars(r)["id"]); err != nil || !id.Valid {
		return pgtype.UUID{}, service.ErrInvalidJobID
	}
	return id, nil
}

func writeSavedSearchError(w http.ResponseWriter, log *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, service.ErrSavedSearchNotFound):
		types.WriteError(w, http.StatusNotFound, "Saved search not found")
	case errors.Is(err, service.ErrSavedSearchNameEmpty):
		types.WriteError(w, http.StatusBadRequest, "Name is required")
	case errors.Is(err, service.ErrSavedSearchNameTaken):
		types.WriteError(w, http.StatusConflict, "A saved search with that name already exists")
	case errors.Is(err, service.ErrSavedSearchInvalid):
		types.WriteError(w, http.StatusBadRequest, "Invalid saved search")
	default:
		log.Errorw("saved search request failed", "error", err)
		types.WriteError(w, http.StatusInternalServerError, "Internal error")
	}
}
