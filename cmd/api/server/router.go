package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephensulimani/internlyapp/cmd/api/middleware"
	"github.com/stephensulimani/internlyapp/cmd/api/routes"
	"go.uber.org/zap"
)

func NewHandler(log *zap.SugaredLogger, pool *pgxpool.Pool) http.Handler {
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.DatabaseMiddleware(pool))
	router.Use(middleware.LoggerContext(log))
	router.PathPrefix("/").Handler(routes.UserRouter())
	return router
}
