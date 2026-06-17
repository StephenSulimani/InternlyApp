package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/types"
	"github.com/stephensulimani/internlyapp/cmd/api/utils"
	"github.com/stephensulimani/internlyapp/internal/db"
)

// hashPassword is swapped in tests to cover hashing failures.
var hashPassword = utils.HashPassword

func UserRouter() *mux.Router {
	router := mux.NewRouter()
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
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, types.ErrorResponse("Error getting request dependencies"), http.StatusInternalServerError)
		return
	}

	parsed_json := RegisterUserBody{}

	err := json.Unmarshal(deps.body, &parsed_json)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, types.ErrorResponse("Error parsing request body"), http.StatusBadRequest)
		return
	}

	if parsed_json.FirstName == "" || parsed_json.LastName == "" || parsed_json.Email == "" || parsed_json.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, types.ErrorResponse("Missing required fields"), http.StatusBadRequest)
		return
	}

	count, err := deps.users.GetUserCount(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, types.ErrorResponse("Error querying the database"), http.StatusInternalServerError)
		deps.log.Error(err)
		return
	}

	hashed_password, err := hashPassword(parsed_json.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, types.ErrorResponse("Error hashing password"), http.StatusInternalServerError)
		deps.log.Error(err)
		return
	}

	false_val := false
	new_user := db.CreateUserParams{
		FirstName: parsed_json.FirstName,
		LastName:  parsed_json.LastName,
		Email:     parsed_json.Email,
		Password:  hashed_password,
		IsActive:  &false_val,
		IsAdmin:   &false_val,
		IsPremium: &false_val,
	}

	if count == 0 {
		true_val := true
		new_user.IsActive = &true_val
		new_user.IsAdmin = &true_val
		new_user.IsPremium = &true_val
	}

	_, err = deps.users.CreateUser(context.Background(), new_user)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, types.ErrorResponse("User already exists"), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, types.ErrorResponse("Error creating user"), http.StatusInternalServerError)
		deps.log.Error(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(types.StringResponse("User successfully registered")))
}
