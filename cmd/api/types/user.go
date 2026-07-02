package types

import (
	"encoding/json"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type UserProfile struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	IsActive  bool   `json:"is_active"`
	IsPremium bool   `json:"is_premium"`
}

type loginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Token   string      `json:"token"`
	User    UserProfile `json:"user"`
}

func UserProfileFrom(u db.User) UserProfile {
	return UserProfile{
		ID:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		IsActive:  u.IsActive,
		IsPremium: u.IsPremium,
	}
}

func WriteLoginSuccess(w http.ResponseWriter, token string, user UserProfile) {
	body, err := json.Marshal(loginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		User:    user,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
