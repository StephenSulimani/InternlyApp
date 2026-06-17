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
