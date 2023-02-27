package service

import "errors"

var (
	// Enumerate possible service errors

	ErrPasswordInvalid        error = errors.New("the password is invalid")
	ErrTokenInvalidExpiration error = errors.New("the token is expired")
	ErrTokenInvalid           error = errors.New("the token is invalid")
	ErrTokenInvalidAudience   error = errors.New("the token audience is invalid")
	ErrTokenInvalidIssuer     error = errors.New("the token issuer is invalid")
	ErrUnsupportedValueType   error = errors.New("the value type is unsupported")
	ErrUsernameInvalid        error = errors.New("the username is invalid")
)
