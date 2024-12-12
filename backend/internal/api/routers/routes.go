package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/plinkplenk/img-share/internal/api/handlers"
	"github.com/plinkplenk/img-share/internal/api/middlewares"
	"github.com/plinkplenk/img-share/internal/auth"
	"github.com/plinkplenk/img-share/internal/users"
	"log/slog"
)

type Opts struct {
	UsersService users.Service
	AuthService  auth.Service
	Logger       *slog.Logger
}

func SetupAPIRouter(parent chi.Router, opts Opts) {
	logger := opts.Logger
	r := chi.NewRouter()
	r.Use(middlewares.Logger(logger))

	authHandler := handlers.NewAuthHandler(opts.AuthService, opts.UsersService, logger)

	r.Mount("/auth", NewAuthRoute(authHandler))

	parent.Mount("/api", r)
}
