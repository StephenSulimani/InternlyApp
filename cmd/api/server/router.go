package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/routes"
	"github.com/stephensulimani/internlyapp/internal/db"
	"github.com/stephensulimani/internlyapp/internal/service"
	"go.uber.org/zap"
)

func NewHandler(log *zap.SugaredLogger, pool *pgxpool.Pool) http.Handler {
	users := service.NewUserService(db.New(pool), nil)

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.DatabaseMiddleware(pool))
	router.Use(middleware.LoggerContext(log))
	router.Use(routes.UserServiceMiddleware(users))
	router.PathPrefix("/").Handler(routes.UserRouter())
	return router
}
