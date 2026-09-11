package routes

import (
	"log/slog"
	"net/http"

	"async/controllers"
	"async/helpers"
	"async/middleware"
	"async/repositories"
)

// New registers the API endpoints and applies request middleware.
func New(controller *controllers.Controller, logger *slog.Logger, jwtManager *helpers.JwtManager, users *repositories.UserStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", controller.Ping)
	mux.HandleFunc("POST /auth/signup", controller.Signup())
	mux.HandleFunc("POST /auth/signin", controller.SignIn())
	mux.HandleFunc("POST /auth/refresh", controller.RefreshToken())
	mux.HandleFunc("POST /reports", controller.RefreshToken())
	mux.HandleFunc("GET /reports/{id}", controller.GetReport())

	chain := middleware.NewLoggerMiddleware(logger)
	chain = middleware.NewAuthMiddleware(jwtManager, users)
	return chain(mux)
}
