package apiserver

import (
	"encoding/json"
	"errors"
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

func (s *ApiServer) signupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingUser, err := s.store.Users.ByEmail(r.Context(), req.Email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	_, err = s.store.Users.CreateUser(r.Context(), req.Password, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ApiResponse[struct{}]{
		Message: "successfully signed up user",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
