package middlewares

import (
	"github.com/go-chi/chi/v5"
	"github.com/plinkplenk/img-share/internal/api"
	"net/http"
)

type writer struct {
	http.ResponseWriter
	code int
}

func (w *writer) WriteHeader(statusCode int) {
	w.code = statusCode
}

func Redirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &writer{ResponseWriter: w}
		redirectValue := chi.URLParam(r, api.RedirectUrlParamName)
		next.ServeHTTP(writer, r)
		if redirectValue != "" {
			http.Redirect(w, r, redirectValue, http.StatusFound)
			return
		}
		w.WriteHeader(writer.code)
	})
}
