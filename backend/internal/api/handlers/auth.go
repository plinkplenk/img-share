package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid/v5"
	"github.com/plinkplenk/img-share/internal/api"
	"github.com/plinkplenk/img-share/internal/auth"
	"github.com/plinkplenk/img-share/internal/users"
	"github.com/plinkplenk/img-share/pkg/cookies"
	"github.com/plinkplenk/img-share/pkg/password"
	"log/slog"
	"net/http"
	"time"
)

type userResponse struct {
	Id        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

type userLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userRegister struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	authService  auth.Service
	usersService users.Service
	logger       *slog.Logger
}

func NewAuthHandler(authService auth.Service, usersService users.Service, logger *slog.Logger) AuthHandler {
	return AuthHandler{
		authService:  authService,
		usersService: usersService,
		logger:       logger,
	}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	ctx := r.Context()
	reader, err := r.GetBody()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			h.logger.Error("cannot close reader", "error", err)
		}
	}()
	userToCreate, err := JSONFromReaderTo[userRegister](reader)
	if err != nil {
		h.logger.Error("cannot unmarshal json", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = h.usersService.GetUserByEmail(ctx, userToCreate.Email)
	if err != nil && !errors.Is(err, users.ErrUserNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		response, err := json.Marshal(BadRequest{Message: "user already exists"})
		if err != nil {
			h.logger.Error("cannot marshal json", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(response); err != nil {
			h.logger.Error("cannot write response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	createdUser, err := h.usersService.CreateUser(
		ctx,
		users.User{
			Email:    userToCreate.Email,
			Password: userToCreate.Password,
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response, err := json.Marshal(BadRequest{Message: "cannot create user"})
		if err != nil {
			h.logger.Error("cannot marshal json", "error", err)
			return
		}
		if _, err := w.Write(response); err != nil {
			h.logger.Error("cannot write response", "error", err)
		}
		return
	}
	response, err := json.Marshal(userResponse{
		Id:        createdUser.Id,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
		IsActive:  createdUser.IsActive,
	})
	if err != nil {
		h.logger.Error("cannot marshal json", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(response); err != nil {
		h.logger.Error("cannot write response", "error", err)
	}
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reader, err := r.GetBody()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			h.logger.Error("cannot close reader", "error", err)
		}
	}()
	userData, err := JSONFromReaderTo[userLogin](reader)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dbUser, err := h.usersService.GetUserByEmail(ctx, userData.Email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !password.Compare(userData.Password, dbUser.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	session, err := h.authService.CreateSession(ctx, dbUser.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     api.SessionIdCookieName,
		Value:    session.Id,
		Path:     "/",
		Expires:  session.ExpiresOn.UTC(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionCookie, err := r.Cookie(api.SessionIdCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := sessionCookie.Valid(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := h.authService.DeleteSessionById(ctx, sessionCookie.Value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cookies.Delete(sessionCookie.Name, w)
	w.WriteHeader(http.StatusOK)
}
