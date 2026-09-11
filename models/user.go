package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id                  uuid.UUID `db:"id"`
	Email               string    `db:"email"`
	HashedPaswordBase64 string    `db:"hashed_password"`
	CreatedAt           time.Time `db:"created_at"`
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
