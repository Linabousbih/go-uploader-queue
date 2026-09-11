package repositories

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"

	"async/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postres"),
	}
}

func (s *UserStore) CreateUser(ctx context.Context, password, email string) (*models.User, error) {
	var user models.User
	const dml = `INSERT INTO USERS (email, hashed_password) VALUES ($1, $2) RETURNING *;`
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password %w", err)
	}
	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(bytes)

	if err := s.db.GetContext(ctx, &user, dml, email, hashedPasswordBase64); err != nil {
		return nil, fmt.Errorf("failed to insert user %w", err)
	}

	return &user, nil
}

func (s *UserStore) ByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	const query = `SELECT 1 FROM USERS WHERE email = $1`
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %w", err)
	}
	return &user, nil
}

func (s *UserStore) ById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	var user models.User
	const query = `SELECT 1 FROM USERS WHERE Id=$1`
	err := s.db.GetContext(ctx, &user, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %w", err)
	}
	return &user, nil
}
