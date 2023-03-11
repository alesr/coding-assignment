package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticateRequest_validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		given    authenticateRequest
		expected error
	}{
		{
			name: "given valid request",
			given: authenticateRequest{
				Username: "foo",
				Password: "bar",
			},
			expected: nil,
		},
		{
			name: "given invalid username",
			given: authenticateRequest{
				Username: "",
				Password: "bar",
			},
			expected: ErrInvalidUsername,
		},
		{
			name: "given invalid password",
			given: authenticateRequest{
				Username: "foo",
				Password: "",
			},
			expected: ErrInvalidPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			observedErr := tc.given.validate()

			assert.Equal(t, tc.expected, observedErr)
		})
	}
}
