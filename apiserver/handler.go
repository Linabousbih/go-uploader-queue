package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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

// [T any] is a place holder to choose the type of data later on
type ApiResponse[T any] struct {
	Data    *T     `json:"data"`
	Message string `json:"message,omitempty"`
}

type SigninResponse struct {
	AccessToken  string
	RefreshToken string
}

// Signup Handler
func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[*CredentialsRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}
		//NewErrWithStatus(...) only creates and returns an error value. It does not write anything to the HTTP response.
		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}
		if existingUser != nil {
			return NewErrWithStatus(http.StatusConflict, fmt.Errorf("user already exists: %w", err))
		}

		_, err = s.store.Users.CreateUser(r.Context(), req.Password, req.Email)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}

		if err := encode(ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}
		return nil
	})

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

// Sign in Handler

func (s *ApiServer) signInHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[*CredentialsRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err := user.ComparePassword(req.Password); err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}
		tokenPair, err := s.jwtManager.GenerateTokenPairs(user.Id)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshToken.Create(r.Context(), user.Id, &tokenPair.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := encode(ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
			Message: "successfully signed in user",
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})
}

// Token Refresh Handler
func (s *ApiServer) tokenrefreshHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[*TokenRefreshRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}

		currentRefrshToken, err := s.jwtManager.Parse(req.RefreshToken)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userIdstr, err := currentRefrshToken.Claims.GetSubject()
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userId, err := uuid.Parse(userIdstr)
		if err != nil {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}
		currentRefrshTokenRecord, err := s.store.RefreshToken.ByPrimaryKey(r.Context(), userId, currentRefrshToken)

		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}
			return NewErrWithStatus(status, err)
		}

		if currentRefrshTokenRecord.ExpiresAt.Before(time.Now()) {
			return NewErrWithStatus(http.StatusUnauthorized, err)
		}

		tokenPair, err := s.jwtManager.GenerateTokenPairs(userId)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		// implement delete
		if _, err := s.store.RefreshToken.Create(r.Context(), userId, &tokenPair.RefreshToken); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		if err = encode(ApiResponse[SigninResponse]{
			Data: &SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})
}
