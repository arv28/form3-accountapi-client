package api

import (
	"fmt"
)

type error interface {
	Error() string
}

type NotFoundError struct {
	ErrorMessage string `json:"error_message"`
}

type ConflictError struct {
	ErrorMessage string `json:"error_message"`
}

type InternalServerError struct {
	ErrorMessage string `json:"error_message"`
}

type BadRequestError struct {
	ErrorMessage string `json:"error_message"`
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not Found Error: %v", e.ErrorMessage)
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error: %v", e.ErrorMessage)
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("Internal Server Error: %v", e.ErrorMessage)
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("Bad Request Error: %v", e.ErrorMessage)
}
