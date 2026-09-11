package models

// [T any] is a place holder to choose the type of data later on
type ApiResponse[T any] struct {
	Data    *T     `json:"data"`
	Message string `json:"message,omitempty"`
}
