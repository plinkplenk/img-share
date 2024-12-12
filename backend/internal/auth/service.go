package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/plinkplenk/img-share/internal/users"
	"log/slog"
	"time"
)

const sessionIdSize = 24

type Service interface {
	CreateSession(ctx context.Context, userId uuid.UUID) (Session, error)
	GetSessionById(ctx context.Context, id string) (Session, error)
	GetSessionsByUserId(ctx context.Context, userId uuid.UUID) ([]Session, error)
	DeleteSessionByUserId(ctx context.Context, userId uuid.UUID, exceptIds ...string) error
	DeleteSessionById(ctx context.Context, id string) error
	GetUserBySessionId(ctx context.Context, id string) (users.User, error)
}

type service struct {
	sessionRepository Repository
	usersRepository   users.Repository
	sessionLifeTime   time.Duration
	timeout           time.Duration
	logger            *slog.Logger
}

func NewService(
	authRepository Repository,
	usersRepository users.Repository,
	sessionLifeTime time.Duration,
	timeout time.Duration, logger *slog.Logger,
) Service {
	return service{
		sessionRepository: authRepository,
		usersRepository:   usersRepository,
		sessionLifeTime:   sessionLifeTime,
		timeout:           timeout,
		logger:            logger,
	}
}

func (s service) generateSessionIdBytes() ([sessionIdSize]byte, error) {
	id := [sessionIdSize]byte{}
	n, err := rand.Read(id[:])
	if err != nil {
		return id, err
	}
	if n != sessionIdSize {
		return id, fmt.Errorf("expected %d bytes, got %d", sessionIdSize, n)
	}
	return id, nil
}

func (s service) sessionIdBytesToHexString(sessionIdBytes [sessionIdSize]byte) string {
	return hex.EncodeToString(sessionIdBytes[:])
}

func (s service) CreateSession(ctx context.Context, userId uuid.UUID) (Session, error) {
	sessionIdBytes, err := s.generateSessionIdBytes()
	if err != nil {
		s.logger.Error("unable to generate session id", "error", err)
		return Session{}, err
	}
	sessionId := s.sessionIdBytesToHexString(sessionIdBytes)
	session := Session{
		Id:        sessionId,
		UserId:    userId,
		ExpiresOn: time.Now().Add(s.sessionLifeTime),
	}
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	createdSession, err := s.sessionRepository.CreateSession(c, session)
	if err != nil {
		s.logger.Error("unable to create session", "error", err)
		return Session{}, err
	}
	return createdSession, nil
}

func (s service) GetSessionById(ctx context.Context, id string) (Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	session, err := s.sessionRepository.GetSessionById(c, id)
	if err != nil {
		s.logger.Error("unable to get session by id", "error", err)
		return Session{}, err
	}
	return session, nil
}
func (s service) GetSessionsByUserId(ctx context.Context, userId uuid.UUID) ([]Session, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	sessions, err := s.sessionRepository.GetSessionsByUserId(c, userId)
	if err != nil {
		s.logger.Error("unable to get sessions", "error", err)
		return []Session{}, err
	}
	return sessions, nil
}

func (s service) DeleteSessionByUserId(ctx context.Context, userId uuid.UUID, exceptIds ...string) error {
	//TODO implement me
	panic("implement me")
}

func (s service) DeleteSessionById(ctx context.Context, sessionId string) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := s.sessionRepository.DeleteSession(c, sessionId); err != nil {
		s.logger.Error("unable to delete session", "error", err)
		return err
	}
	return nil
}

func (s service) GetUserBySessionId(ctx context.Context, sessionId string) (users.User, error) {
	// TODO maybe implement this in auth repository
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	session, err := s.sessionRepository.GetSessionById(c, sessionId)
	if err != nil {
		return users.User{}, err
	}
	return s.usersRepository.GetUserById(c, session.UserId)
}
