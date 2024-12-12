package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/plinkplenk/img-share/internal/api/handlers"
	"github.com/plinkplenk/img-share/internal/api/middlewares"
)

func NewAuthRoute(handler handlers.AuthHandler) chi.Router {
	r := chi.NewRouter()
	r.With(middlewares.Redirect).Post("/sign-up", handler.Register)
	r.With(middlewares.Redirect).Post("/sign-in", handler.Login)
	r.Post("/sign-out", handler.Logout)
	return r
}
