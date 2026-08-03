package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (s *RefreshTokenStore) Create(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const insert = `INSERT INTO refresh_tokens(user_id, hashed_token, expires_at,) VALUES ($1, $2, $3) RETURNING *;`
	hashed_token, err := bcrypt.GenerateFromPassword([]byte(token.Raw), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	expiresAt, err := token.Claims.GetExpirationTime()

	var refreshTokenHash RefreshToken
	if err := s.db.GetContext(ctx, &refreshTokenHash, insert, userId, hashed_token, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to create refresh token record %w", err)
	}
	return &refreshTokenHash, nil
}

func (s *RefreshTokenStore) ByPrimaryKey(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const query = `SELECT * FROM refresh_tokens WHERE user_id=$1 AND hashed_token=$2;`
	hashed_token, err := bcrypt.GenerateFromPassword([]byte(token.Raw), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, query, userId, hashed_token); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &refreshToken, nil
}
