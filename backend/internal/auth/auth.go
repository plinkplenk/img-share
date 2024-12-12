package auth

import (
	"context"
	"github.com/gofrs/uuid/v5"
	"time"
)

type Session struct {
	Id        string
	UserId    uuid.UUID
	ExpiresOn time.Time
}

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	GetSessionById(ctx context.Context, value string) (Session, error)
	GetSessionsByUserId(ctx context.Context, userId uuid.UUID) ([]Session, error)
	DeleteSession(ctx context.Context, sessionValue string) error
	DeleteSessionsByUserId(ctx context.Context, userId uuid.UUID, exceptIds ...string) error
}
