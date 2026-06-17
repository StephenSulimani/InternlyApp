package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/service"
	"golang.org/x/time/rate"
)

func UserRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.RateLimit(rate.Every(time.Minute/5), 5))
	router.Use(middleware.EnsureJSONBody)
	router.HandleFunc("/register", RegisterUser).Methods("POST")
	return router
}

type RegisterUserBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	deps, ok := depsFromRequest(w, r)
	if !ok {
		http.Error(w, types.ErrorResponse("Error getting request dependencies"), http.StatusInternalServerError)
		return
	}

	var body RegisterUserBody
	if err := json.Unmarshal(deps.body, &body); err != nil {
		http.Error(w, types.ErrorResponse("Error parsing request body"), http.StatusBadRequest)
		return
	}

	err := deps.users.Register(r.Context(), service.RegisterInput{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Password:  body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingFields):
			http.Error(w, types.ErrorResponse("Missing required fields"), http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidEmail):
			http.Error(w, types.ErrorResponse("Invalid email address"), http.StatusBadRequest)
		case errors.Is(err, service.ErrWeakPassword):
			http.Error(w, types.ErrorResponse("Password must be at least 8 characters"), http.StatusBadRequest)
		case errors.Is(err, service.ErrUserExists):
			http.Error(w, types.ErrorResponse("User already exists"), http.StatusBadRequest)
		case errors.Is(err, service.ErrCountUsers):
			deps.log.Error(err)
			http.Error(w, types.ErrorResponse("Error querying the database"), http.StatusInternalServerError)
		case errors.Is(err, service.ErrHashPassword):
			deps.log.Error(err)
			http.Error(w, types.ErrorResponse("Error hashing password"), http.StatusInternalServerError)
		case errors.Is(err, service.ErrCreateUser):
			deps.log.Error(err)
			http.Error(w, types.ErrorResponse("Error creating user"), http.StatusInternalServerError)
		default:
			deps.log.Error(err)
			http.Error(w, types.ErrorResponse("Error creating user"), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(types.StringResponse("User successfully registered")))
}
