package handlers

import (
	"encoding/json"
	"errors"
	"github.com/plinkplenk/img-share/internal/api"
	"github.com/plinkplenk/img-share/internal/auth"
	"github.com/plinkplenk/img-share/internal/users"
	"io"
	"net/http"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type BadRequest struct {
	Message string `json:"message"`
}

func JSONFromReaderTo[T any](reader io.ReadCloser) (T, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return *new(T), err
	}
	return JSONFromBytesTo[T](bytes)
}

func JSONFromBytesTo[T any](data []byte) (T, error) {
	target := *new(T)
	if err := json.Unmarshal(data, &target); err != nil {
		return target, err
	}
	return target, nil
}

func GetUserFromSession(authService auth.Service, r *http.Request) (users.User, error) {
	sessionId, err := r.Cookie(api.SessionIdCookieName)
	if err != nil {
		return users.User{}, ErrUnauthorized
	}
	user, err := authService.GetUserBySessionId(r.Context(), sessionId.Value)
	if err != nil {
		return users.User{}, err
	}
	return user, nil
}
