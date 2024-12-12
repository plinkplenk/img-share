package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const postgresRepositorySource = "users.repo.pg"

// TODO: maybe this shouldn't be here
var allowedUpdateFields = map[string]struct{}{
	"email":     {},
	"is_active": {},
	"password":  {},
}

type postgresRepository struct {
	db *pgxpool.Pool
}

type pgUser struct {
	id        pgtype.UUID
	email     string
	password  string
	isActive  bool
	createdAt pgtype.Timestamp
}

func fromPGUser(user pgUser) (User, error) {
	id, err := uuid.FromBytes(user.id.Bytes[:])
	if err != nil {
		return User{}, err
	}
	return User{
		Id:        id,
		Email:     user.email,
		Password:  user.password,
		IsActive:  user.isActive,
		CreatedAt: user.createdAt.Time,
	}, nil
}

func toPGUser(user User) pgUser {
	return pgUser{
		id: pgtype.UUID{
			Bytes: [16]byte(user.Id.Bytes()),
		},
		email:     user.Email,
		password:  user.Password,
		isActive:  user.IsActive,
		createdAt: pgtype.Timestamp{Time: user.CreatedAt.UTC()},
	}
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{
		db,
	}
}

func (r *postgresRepository) getByField(ctx context.Context, field string, value any) (User, error) {
	const op = postgresRepositorySource + ".getByField"
	query := fmt.Sprintf(
		`SELECT id, email, password, is_active, created_at FROM users WHERE %s = $1`,
		field,
	)
	row := r.db.QueryRow(ctx, query, value)
	var user pgUser
	if err := row.Scan(
		&user.id,
		&user.email,
		&user.password,
		&user.isActive,
		&user.createdAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("[%s]: %e", op, err)
	}
	return fromPGUser(user)
}

func (r *postgresRepository) GetUserById(ctx context.Context, id uuid.UUID) (User, error) {
	return r.getByField(ctx, "id", id)
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return r.getByField(ctx, "email", email)
}

func (r *postgresRepository) UpdateUser(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (User, error) {
	const op = postgresRepositorySource + ".UpdateUser"
	if len(fieldValue) == 0 {
		return r.GetUserById(ctx, id)
	}
	args := make([]any, len(fieldValue))
	fieldsToSet := make([]string, len(fieldValue))
	for field, value := range fieldValue {
		// TODO: think about it later
		if _, ok := allowedUpdateFields[field]; !ok {
			continue
		}
		args = append(args, value)
		fieldsToSet = append(
			fieldsToSet,
			fmt.Sprintf("%s = $%d, ", field, len(args)),
		)
	}
	if len(fieldsToSet) == 0 {
		return r.GetUserById(ctx, id)
	}
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d`,
		strings.Join(fieldsToSet, ", "),
		len(args),
	)

	row := r.db.QueryRow(ctx, query, args...)
	var user pgUser
	if err := row.Scan(&user.id, &user.email, &user.password, &user.isActive, &user.createdAt); err != nil {
		return User{}, fmt.Errorf("[%s]: %e", op, err)
	}
	return fromPGUser(user)
}

func (r *postgresRepository) CreateUser(ctx context.Context, user User) (User, error) {
	const op = postgresRepositorySource + ".CreateSession"
	query := `
INSERT INTO users (email, password, is_active) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id, email, password, is_active, created_at
`
	row := r.db.QueryRow(ctx, query, user.Email, user.Password, user.IsActive)
	var createdUser User
	if err := row.Scan(
		&createdUser.Id,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.IsActive,
		&createdUser.CreatedAt,
	); err != nil {
		return user, fmt.Errorf("[%s]: %e", op, err)
	}
	return user, nil
}

func (r *postgresRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = postgresRepositorySource + ".DeleteUser"
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return fmt.Errorf("[%s]: %e", op, err)
}
