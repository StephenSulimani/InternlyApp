package types

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type SavedSearch struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Q         string `json:"q"`
	Type      string `json:"type"`
	Location  string `json:"location"`
	Recency   string `json:"recency"`
	Saved     bool   `json:"saved"`
	Sort      string `json:"sort"`
	Order     string `json:"order"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func SavedSearchFrom(row db.UserSavedSearch) SavedSearch {
	search := SavedSearch{
		ID:       row.ID.String(),
		Name:     row.Name,
		Q:        row.Q,
		Type:     row.JobType,
		Location: row.Location,
		Recency:  row.Recency,
		Saved:    row.SavedOnly,
		Sort:     row.SortBy,
		Order:    row.SortDir,
	}
	if row.CreatedAt.Valid {
		search.CreatedAt = row.CreatedAt.Time.UTC().Format(time.RFC3339)
	}
	if row.UpdatedAt.Valid {
		search.UpdatedAt = row.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}
	return search
}

func SavedSearchesFrom(rows []db.UserSavedSearch) []SavedSearch {
	out := make([]SavedSearch, 0, len(rows))
	for _, row := range rows {
		out = append(out, SavedSearchFrom(row))
	}
	return out
}

type savedSearchesResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    []SavedSearch `json:"data"`
}

func WriteSavedSearches(w http.ResponseWriter, searches []SavedSearch) {
	if searches == nil {
		searches = []SavedSearch{}
	}
	body, err := json.Marshal(savedSearchesResponse{
		Success: true,
		Message: "Saved searches retrieved",
		Data:    searches,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

type savedSearchResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    SavedSearch `json:"data"`
}

func WriteSavedSearch(w http.ResponseWriter, status int, message string, search SavedSearch) {
	body, err := json.Marshal(savedSearchResponse{
		Success: true,
		Message: message,
		Data:    search,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
