package routes

import (
	"errors"
	"net/http"

	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

func writeRegisterError(w http.ResponseWriter, log *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, service.ErrMissingFields):
		types.WriteError(w, http.StatusBadRequest, "Missing required fields")
	case errors.Is(err, service.ErrInvalidEmail):
		types.WriteError(w, http.StatusBadRequest, "Invalid email address")
	case errors.Is(err, service.ErrWeakPassword):
		types.WriteError(w, http.StatusBadRequest, "Password must be at least 8 characters")
	case errors.Is(err, service.ErrUserExists):
		types.WriteError(w, http.StatusConflict, "User already exists")
	case errors.Is(err, service.ErrCountUsers):
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error querying the database")
	case errors.Is(err, service.ErrHashPassword):
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error hashing password")
	case errors.Is(err, service.ErrCreateUser):
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error creating user")
	default:
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error creating user")
	}
}

func writeLoginError(w http.ResponseWriter, log *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, service.ErrMissingFields):
		types.WriteError(w, http.StatusBadRequest, "Missing required fields")
	case errors.Is(err, service.ErrInvalidEmailOrPassword):
		types.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
	case errors.Is(err, service.ErrUserInactive):
		types.WriteError(w, http.StatusForbidden, "Account pending activation")
	case errors.Is(err, service.ErrGetUser):
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error querying the database")
	default:
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error logging in")
	}
}

func writeJobsError(w http.ResponseWriter, log *zap.SugaredLogger, err error) {
	switch {
	case errors.Is(err, service.ErrGetJobs):
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error querying the database")
	default:
		log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error retrieving jobs")
	}
}
