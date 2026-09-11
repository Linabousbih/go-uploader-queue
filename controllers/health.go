package controllers

import (
	"net/http"
)

func (s *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {

	}
}
