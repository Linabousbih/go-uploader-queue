package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sqlx.DB
}

type User struct {
	Id                  uuid.UUID `db:"id"`
	Email               string    `db:"email"`
	HashedPaswordBase64 string    `db:"hashed_password"`
	CreatedAt           time.Time `db:"created_at"`
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: sqlx.NewDb(db, "postres"),
	}
}

func (u *User) ComparePassword(password string) error {
	bytes, err := base64.StdEncoding.DecodeString(u.HashedPaswordBase64)
	if err != nil {
		return fmt.Errorf("failed decoding password %w", err)
	}
	err = bcrypt.CompareHashAndPassword(bytes, []byte(password))
	if err != nil {
		return fmt.Errorf("invalid passowrd %w", err)
	}

	return nil
}

func (s *UserStore) CreateUser(ctx context.Context, password, email string) (*User, error) {
	var user User
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

func (s *UserStore) ByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	const query = `SELECT 1 FROM USERS WHERE email = $1`
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %w", err)
	}
	return &user, nil
}

func (s *UserStore) ById(ctx context.Context, userId uuid.UUID) (*User, error) {
	var user User
	const query = `SELECT 1 FROM USERS WHERE Id=$1`
	err := s.db.GetContext(ctx, &user, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %w", err)
	}
	return &user, nil
}
