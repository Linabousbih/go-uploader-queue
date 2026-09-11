package models

import (
	"errors"
)

type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *CredentialsRequest) Validate() error {
	if s.Email == "" {
		return errors.New("email is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

type SigninResponse struct {
	AccessToken  string
	RefreshToken string
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *TokenRefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}
