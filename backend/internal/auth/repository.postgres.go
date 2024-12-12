package auth

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const postgresRepositorySource = "auth.repo.pg"

type postgresRepository struct {
	db *pgxpool.Pool
}

type pgSession struct {
	id        string
	userId    pgtype.UUID
	expiresOn pgtype.Timestamp
}

func fromPGSession(session pgSession) (Session, error) {
	userId, err := uuid.FromBytes(session.userId.Bytes[:])
	if err != nil {
		return Session{}, err
	}
	return Session{
		Id:        session.id,
		UserId:    userId,
		ExpiresOn: session.expiresOn.Time,
	}, nil
}

func toPGSession(session Session) pgSession {
	return pgSession{
		id: session.Id,
		userId: pgtype.UUID{
			Bytes: [16]byte(session.UserId.Bytes()),
		},
		expiresOn: pgtype.Timestamp{
			Time: session.ExpiresOn.UTC(),
		},
	}
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) CreateSession(ctx context.Context, session Session) (Session, error) {
	const op = postgresRepositorySource + ".CreateSession"
	query := `
INSERT INTO auth_sessions
	(id, user_id, expires_on) 
	VALUES ($1, $2, $3)
	RETURNING id, user_id, expires_on;
`

	sessionToCreate := toPGSession(session)
	var createdSession pgSession
	if err := r.db.QueryRow(ctx, query, sessionToCreate.id, sessionToCreate.userId, sessionToCreate.expiresOn).Scan(
		&createdSession.id,
		&createdSession.userId,
		&createdSession.expiresOn,
	); err != nil {
		return Session{}, fmt.Errorf("[%s]: %w", op, err)
	}
	return fromPGSession(createdSession)
}

func (r *postgresRepository) getSessionsByField(ctx context.Context, field string, value any) ([]Session, error) {
	const op = postgresRepositorySource + ".getSessionsByField"
	query := fmt.Sprintf(
		`
SELECT id, user_id, expires_on FROM auth_sessions WHERE %s = $1`,
		field,
	)
	var sessions []Session
	rows, err := r.db.Query(ctx, query, value)
	defer rows.Close()
	if err != nil {
		return []Session{}, fmt.Errorf("[%s]: %w", op, err)
	}
	for rows.Next() {
		if rows.Err() != nil {
			return []Session{}, rows.Err()
		}
		var pgSession pgSession
		if err := rows.Scan(&pgSession.id, &pgSession.userId, &pgSession.expiresOn); err != nil {
			return []Session{}, err
		}
		session, err := fromPGSession(pgSession)
		if err != nil {
			return []Session{}, fmt.Errorf("[%s]: %w", op, err)
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *postgresRepository) GetSessionById(ctx context.Context, value string) (Session, error) {
	const op = postgresRepositorySource + ".GetSessionById"
	sessions, err := r.getSessionsByField(ctx, "value", value)
	if err != nil {
		return Session{}, fmt.Errorf("[%s]: %e", op, err)
	}
	return sessions[0], nil
}

func (r *postgresRepository) GetSessionsByUserId(ctx context.Context, userId uuid.UUID) ([]Session, error) {
	return r.getSessionsByField(ctx, "user_id", userId)
}

func (r *postgresRepository) DeleteSession(ctx context.Context, sessionValue string) error {
	const op = postgresRepositorySource + ".DeleteSession"
	query := `DELETE FROM auth_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionValue)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	return nil
}

func (r *postgresRepository) DeleteSessionsByUserId(ctx context.Context, userId uuid.UUID, exceptValues ...string) error {
	const op = postgresRepositorySource + ".DeleteSessionsByUserId"
	query := `DELETE FROM auth_sessions WHERE user_id = $1`
	if len(exceptValues) > 0 {
		// adding placeholders
		placeholders := make([]string, len(exceptValues))
		for i := 1; i <= len(exceptValues); i++ {
			placeholders[i-1] = fmt.Sprintf("$%d", i)
		}
		query = fmt.Sprintf("[%s] AND value NOT IN (%s)", query, strings.Join(placeholders, ", "))
	}
	_, err := r.db.Exec(ctx, query, userId, exceptValues)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	return nil
}

func (r *postgresRepository) DeleteExpiredSessions(ctx context.Context) error {
	const op = postgresRepositorySource + ".DeleteExpiredSessions"
	query := `DELETE FROM auth_sessions WHERE expires_on < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	return nil
}
