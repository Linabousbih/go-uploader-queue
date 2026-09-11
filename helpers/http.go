package helpers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"async/models"
)

type ErrWithStatus struct {
	status int
	err    error
}

func (e *ErrWithStatus) Error() string {
	return e.err.Error()
}

func NewErrWithStatus(status int, err error) *ErrWithStatus {
	return &ErrWithStatus{
		status: status,
		err:    err,
	}
}

// We are customizing what errors would return a status code, what errors are logged and what errors are shown to the user
func Handler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// core Handler is called here
		if err := f(w, r); err != nil {
			// Default status in case the Handler did not receive a ErrWithStatus type
			status := http.StatusInternalServerError
			msg := http.StatusText(status)
			if e, ok := err.(*ErrWithStatus); ok {
				status = e.status
				msg = http.StatusText(e.status)
				if status == http.StatusBadRequest || status == http.StatusConflict {
					msg = e.err.Error()
				}
			}

			slog.Error("error executing Handler", "error", err, "status", status, "message", msg)
			w.WriteHeader(status)

			// if err := json.NewEncoder(w).Encode(models.ApiResponse[struct{}]{
			// 	Message: msg,
			// }); err != nil {
			// 	slog.Error("error encoding response", "error", err)
			// }
			if err := Encode(models.ApiResponse[struct{}]{
				Message: msg,
			}, status, w); err != nil {
				slog.Error("error encoding response", "error", err)
			}
		}
	}
}

func Encode[T any](v T, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encoding response %w", err)
	}

	return nil
}

type Validator interface {
	Validate() error
}

func Decode[T Validator](r *http.Request) (T, error) {
	var t T
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return t, fmt.Errorf("decoding request body: %w", err)
	}
	if err := t.Validate(); err != nil {
		return t, err
	}

	return t, nil
}
