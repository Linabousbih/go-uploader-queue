package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"async/helpers"
	"async/models"

	"github.com/google/uuid"
)

// Signup handles user registration.
func (s *Controller) Signup() http.HandlerFunc {
	return helpers.Handler(func(w http.ResponseWriter, r *http.Request) error {

		req, err := helpers.Decode[*models.CredentialsRequest](r)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusBadRequest, err)
		}
		//helpers.NewErrWithStatus(...) only creates and returns an error value. It does not write anything to the HTTP response.
		existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}
		if existingUser != nil {
			return helpers.NewErrWithStatus(http.StatusConflict, fmt.Errorf("user already exists: %w", err))
		}

		_, err = s.store.Users.CreateUser(r.Context(), req.Password, req.Email)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}

		if err := helpers.Encode(models.ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("invalid request: %w", err))
		}
		return nil
	})

}

// SignIn exchanges credentials for a token pair.

func (s *Controller) SignIn() http.HandlerFunc {
	return helpers.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helpers.Decode[*models.CredentialsRequest](r)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusBadRequest, err)
		}

		user, err := s.store.Users.ByEmail(r.Context(), req.Email)
		if err := user.ComparePassword(req.Password); err != nil {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}
		tokenPair, err := s.jwtManager.GenerateTokenPairs(user.Id)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.store.RefreshToken.Create(r.Context(), user.Id, &tokenPair.RefreshToken)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := helpers.Encode(models.ApiResponse[models.SigninResponse]{
			Data: &models.SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
			Message: "successfully signed in user",
		}, http.StatusOK, w); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})
}

// RefreshToken issues a new token pair.
func (s *Controller) RefreshToken() http.HandlerFunc {
	return helpers.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helpers.Decode[*models.TokenRefreshRequest](r)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusBadRequest, err)
		}

		currentRefrshToken, err := s.jwtManager.Parse(req.RefreshToken)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userIdstr, err := currentRefrshToken.Claims.GetSubject()
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}

		userId, err := uuid.Parse(userIdstr)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}
		currentRefrshTokenRecord, err := s.store.RefreshToken.ByPrimaryKey(r.Context(), userId, currentRefrshToken)

		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, sql.ErrNoRows) {
				status = http.StatusUnauthorized
			}
			return helpers.NewErrWithStatus(status, err)
		}

		if currentRefrshTokenRecord.ExpiresAt.Before(time.Now()) {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}

		tokenPair, err := s.jwtManager.GenerateTokenPairs(userId)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		// implement delete
		if _, err := s.store.RefreshToken.Create(r.Context(), userId, &tokenPair.RefreshToken); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		if err = helpers.Encode(models.ApiResponse[models.SigninResponse]{
			Data: &models.SigninResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})
}
