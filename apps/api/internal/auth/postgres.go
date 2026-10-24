package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polarky2026/kyno/apps/api/internal/db"
)

const uniqueViolation = "23505"

type PostgresStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool, queries: db.New(pool)}
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return toUser(row), nil
}

func (s *PostgresStore) CreateUserWithVerificationToken(ctx context.Context, u NewUser, tokenHash []byte, expiresAt time.Time) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)

		user, err := q.CreateUser(ctx, db.CreateUserParams{
			Email:        u.Email,
			Name:         u.Name,
			PasswordHash: &u.PasswordHash,
		})
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		if err != nil {
			return err
		}

		return q.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
			TokenHash: tokenHash,
			UserID:    user.ID,
			ExpiresAt: expiresAt,
		})
	})
}

func (s *PostgresStore) ReplaceVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)

		if err := q.DeleteEmailVerificationTokensForUser(ctx, userID); err != nil {
			return err
		}

		return q.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
			TokenHash: tokenHash,
			UserID:    userID,
			ExpiresAt: expiresAt,
		})
	})
}

func (s *PostgresStore) ConsumeVerificationToken(ctx context.Context, tokenHash []byte) (uuid.UUID, time.Time, error) {
	row, err := s.queries.ConsumeEmailVerificationToken(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, time.Time{}, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}
	return row.UserID, row.ExpiresAt, nil
}

func (s *PostgresStore) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	return s.queries.MarkEmailVerified(ctx, userID)
}

func toUser(row db.User) User {
	return User{
		ID:              row.ID,
		Email:           row.Email,
		Name:            row.Name,
		PasswordHash:    row.PasswordHash,
		EmailVerifiedAt: row.EmailVerifiedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
