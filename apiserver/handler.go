package apiserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *SignupRequest) Validate() error {
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

func (s *ApiServer) signupHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := decode[SignupRequest](r)
		if err != nil {
			return NewErrWithStatus(http.StatusBadRequest, err)
		}
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
