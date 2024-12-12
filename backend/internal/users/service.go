package users

import (
	"context"
	"errors"
	"github.com/gofrs/uuid/v5"
	"github.com/plinkplenk/img-share/pkg/password"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Service interface {
	GetUserById(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, newPassword, oldPassword string) error
	CreateUser(ctx context.Context, user User) (User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

var (
	ErrPasswordsDidNotMatch = errors.New("passwords did not match")
)

type service struct {
	repository Repository
	logger     *slog.Logger
	timeout    time.Duration
}

func NewService(repository Repository, timeout time.Duration, logger *slog.Logger) Service {
	return service{
		repository: repository,
		timeout:    timeout,
		logger:     logger,
	}
}

func (s service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s service) GetUserById(ctx context.Context, id uuid.UUID) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	user, err := s.repository.GetUserById(c, id)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.logger.Error("cannot get user by id", "error", err)
	}
	return user, err
}

func (s service) GetUserByEmail(ctx context.Context, email string) (User, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	user, err := s.repository.GetUserByEmail(c, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.logger.Error("cannot get user by email", "error", err)
	}
	return user, err
}

func (s service) ChangePassword(ctx context.Context, id uuid.UUID, newPassword, oldPassword string) error {
	c1, cancel1 := context.WithTimeout(ctx, s.timeout)
	defer cancel1()
	user, err := s.GetUserById(c1, id)
	if err != nil {
		s.logger.Error("cannot get user by id", "error", err)
		return err
	}
	if !password.Compare(user.Password, oldPassword) {
		return ErrPasswordsDidNotMatch
	}
	hash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}
	c2, cancel2 := context.WithTimeout(ctx, s.timeout)
	defer cancel2()
	if _, err := s.repository.UpdateUser(c2, id, map[string]any{"password": hash}); err != nil {
		s.logger.Error("cannot update user", "error", err)
		return err
	}
	return nil
}

func (s service) CreateUser(ctx context.Context, user User) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	user.Id = uuid.Must(uuid.NewV4())
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now().UTC()
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	createdUser, err := s.repository.CreateUser(c, user)
	if err != nil {
		s.logger.Error("cannot create user", "error", err)
		return User{}, err
	}
	return createdUser, err
}

func (s service) UpdateUser(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (User, error) {
	//TODO implement me
	panic("implement me")
}

func (s service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.repository.DeleteUser(c, id); err != nil {
		s.logger.Error("cannot delete user", "error", err)
		return err
	}
	return nil
}
