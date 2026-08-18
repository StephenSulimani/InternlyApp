package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/internal/service"
	"golang.org/x/time/rate"
)

func APIRouter() *mux.Router {
	router := mux.NewRouter()
	// Board browsing fires several GETs (jobs, locations, stats) plus
	// search/sort/pagination. Keep this high enough for interactive use.
	router.Use(middleware.RateLimit(5, 30))

	// Public
	router.HandleFunc("/board/preview", BoardPreview).Methods("GET")
	router.HandleFunc("/jobs/stats", JobStats).Methods("GET")

	jsonRouter := router.NewRoute().Subrouter()
	jsonRouter.Use(middleware.RateLimit(rate.Every(time.Minute/5), 5))
	jsonRouter.Use(middleware.EnsureJSONBody)
	jsonRouter.HandleFunc("/register", RegisterUser).Methods("POST")
	jsonRouter.HandleFunc("/login", LoginUser).Methods("POST")

	// Authenticated members only
	authed := router.NewRoute().Subrouter()
	authed.Use(RequireAuth)
	authed.HandleFunc("/jobs", ListJobs).Methods("GET")
	authed.HandleFunc("/jobs/locations", JobLocations).Methods("GET")

	return router
}

type RegisterUserBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	deps, ok := depsFromRequest(w, r)
	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	var body RegisterUserBody
	if err := json.Unmarshal(deps.body, &body); err != nil {
		types.WriteError(w, http.StatusBadRequest, "Error parsing request body")
		return
	}

	err := deps.users.Register(r.Context(), service.RegisterInput{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Password:  body.Password,
	})
	if err != nil {
		writeRegisterError(w, deps.log, err)
		return
	}

	types.WriteSuccess(w, http.StatusCreated, "User successfully registered")
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	deps, ok := depsFromRequest(w, r)

	if !ok {
		types.WriteError(w, http.StatusInternalServerError, "Error getting request dependencies")
		return
	}

	var body LoginUserBody
	if err := json.Unmarshal(deps.body, &body); err != nil {
		types.WriteError(w, http.StatusBadRequest, "Error parsing request body")
		return
	}

	user, err := deps.users.Login(r.Context(), service.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeLoginError(w, deps.log, err)
		return
	}

	token, err := deps.tokens.Issue(user)
	if err != nil {
		deps.log.Error(err)
		types.WriteError(w, http.StatusInternalServerError, "Error issuing token")
		return
	}

	types.WriteLoginSuccess(w, token, types.UserProfileFrom(user))
}
