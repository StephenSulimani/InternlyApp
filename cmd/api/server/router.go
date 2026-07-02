package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/routes"
	"github.com/stephensulimani/internlyapp/internal/auth"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"github.com/stephensulimani/internlyapp/internal/utils"
	"go.uber.org/zap"
)

func NewHandler(log *zap.SugaredLogger, pool *pgxpool.Pool) http.Handler {
	queries := db.New(pool)
	users := service.NewUserService(queries, nil)
	jobs := service.NewJobService(queries)

	jwtSecret := utils.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	ttlHours, _ := strconv.Atoi(utils.GetEnv("JWT_TTL_HOURS", "168"))
	if ttlHours <= 0 {
		ttlHours = 168
	}
	tokens := auth.NewTokenIssuer(jwtSecret, time.Duration(ttlHours)*time.Hour)

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.DatabaseMiddleware(pool))
	router.Use(middleware.LoggerContext(log))
	router.Use(routes.UserServiceMiddleware(users))
	router.Use(routes.JobServiceMiddleware(jobs))
	router.Use(routes.TokenIssuerMiddleware(tokens))
	router.PathPrefix("/").Handler(routes.APIRouter())
	return router
}
