package users

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	Id        uuid.UUID
	Email     string
	Password  string
	IsActive  bool
	CreatedAt time.Time
}

type Repository interface {
	GetUserById(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, user User) (User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
