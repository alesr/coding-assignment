package app

import (
	"errors"
	"net/http"

	"github.com/alesr/code-assignment/internal/service"
)

type APIError struct {
	StatusCode  int    `json:"status_code"`
	Description string `json:"error"`
}

// Implement the error interface.
func (e APIError) Error() string {
	return e.Description
}

var (
	// Enumerate possible transport errors.

	ErrInvalidRequest = APIError{
		StatusCode:  http.StatusBadRequest,
		Description: "the request is invalid",
	}

	ErrInvalidUsername = APIError{
		StatusCode:  http.StatusBadRequest,
		Description: "the username is invalid",
	}

	ErrInvalidPassword = APIError{
		StatusCode:  http.StatusBadRequest,
		Description: "the password is invalid",
	}

	ErrUnsupportedValueType = APIError{
		StatusCode:  http.StatusUnprocessableEntity,
		Description: "the value type is unsupported",
	}

	ErrUnauthorized = APIError{
		StatusCode:  http.StatusUnauthorized,
		Description: "unauthorized",
	}

	ErrInternal = APIError{
		StatusCode:  http.StatusInternalServerError,
		Description: "internal server error",
	}
)

// translate service errors into transport errors.
func toTransportError(err error) error {
	if errors.Is(err, service.ErrUsernameInvalid) {
		return ErrInvalidUsername
	}

	if errors.Is(err, service.ErrPasswordInvalid) {
		return ErrInvalidPassword
	}

	if errors.Is(err, service.ErrTokenInvalid) {
		return ErrUnauthorized
	}

	if errors.Is(err, service.ErrTokenInvalidExpiration) {
		return ErrUnauthorized
	}

	if errors.Is(err, service.ErrTokenInvalidIssuer) {
		return ErrUnauthorized
	}

	if errors.Is(err, service.ErrTokenInvalidAudience) {
		return ErrUnauthorized
	}

	if errors.Is(err, service.ErrUnsupportedValueType) {
		return ErrUnsupportedValueType
	}
	return ErrInternal
}
