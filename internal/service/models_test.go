package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentials_validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		credentials   Credentials
		expectedError error
	}{
		{
			name: "valid credentials",
			credentials: Credentials{
				Username: "foo",
				Password: "bar",
			},
			expectedError: nil,
		},
		{
			name: "invalid username",
			credentials: Credentials{
				Username: "",
				Password: "bar",
			},
			expectedError: ErrUsernameInvalid,
		},
		{
			name: "invalid password",
			credentials: Credentials{
				Username: "foo",
				Password: "",
			},
			expectedError: ErrPasswordInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			observedErr := tc.credentials.validate()
			assert.True(t, errors.Is(observedErr, tc.expectedError))
		})
	}
}
