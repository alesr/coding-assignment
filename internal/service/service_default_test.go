package service

import (
	"context"
	"errors"
	"testing"

	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultService_GenerateToken(t *testing.T) {
	t.Parallel()

	creds := Credentials{
		Username: "foo-username",
		Password: "bar-password",
	}

	logger := zap.NewNop()
	jwtKey := []byte("foo-key")

	service := NewDefaultService(logger, jwtKey)

	observedToken, err := service.GenerateToken(context.TODO(), creds)
	require.NoError(t, err)

	assert.NotEmpty(t, observedToken.AccessToken)
	assert.Equal(t, "Bearer", observedToken.TokenType)
	assert.True(t, observedToken.ExpiresIn > 0)

	// Parse the token so we can verify the claim values.
	parsedTkn, err := jwt.ParseWithClaims(
		observedToken.AccessToken,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		},
	)
	require.NoError(t, err)

	claims, ok := parsedTkn.Claims.(*Claims)
	require.True(t, ok)

	assert.Equal(t, creds.Username, claims.Subject)
	assert.Equal(t, "qredo", claims.Issuer)
	assert.Equal(t, "qredo", claims.Audience)
	assert.True(t, time.Now().Before(time.Unix(claims.ExpiresAt, 0)))
}

func TestDefaultService_Authenticate_InvalidCredentials(t *testing.T) {
	t.Parallel()

	service := NewDefaultService(zap.NewNop(), nil)

	t.Run("invalid username", func(t *testing.T) {
		givenCreds := Credentials{
			Username: "",
			Password: "bar",
		}

		_, err := service.GenerateToken(context.TODO(), givenCreds)
		assert.Equal(t, ErrUsernameInvalid, errors.Unwrap(err))
	})

	t.Run("invalid password", func(t *testing.T) {
		givenCreds := Credentials{
			Username: "foo",
			Password: "",
		}

		_, err := service.GenerateToken(context.TODO(), givenCreds)
		assert.Equal(t, ErrPasswordInvalid, errors.Unwrap(err))
	})
}

func TestDefaultService_VerifyToken(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop()
	jwtKey := []byte("test-key")
	username := "foo-username"

	service := NewDefaultService(logger, jwtKey)

	t.Run("valid token", func(t *testing.T) {
		givenCreds := Credentials{
			Username: username,
			Password: "bar",
		}

		givenToken, err := service.GenerateToken(context.TODO(), givenCreds)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), givenToken.AccessToken)
		assert.NoError(t, observedErr)
	})

	t.Run("expired token", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:   username,
				Issuer:    "qredo",
				Audience:  "qredo",
				ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.Error(t, observedErr)
	})

	t.Run("invalid issuer", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:   username,
				Issuer:    "invalid-issuer",
				Audience:  "qredo",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.Error(t, observedErr)
	})

	t.Run("issuer is not set", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:   username,
				Audience:  "qredo",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.True(t, errors.Is(observedErr, ErrTokenInvalidIssuer))
	})

	t.Run("invalid audience", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:   username,
				Issuer:    "qredo",
				Audience:  "invalid-audience",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.Error(t, observedErr)
	})

	t.Run("audience is not set", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:   username,
				Issuer:    "qredo",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.True(t, errors.Is(observedErr, ErrTokenInvalidAudience))
	})

	t.Run("expiration is not set", func(t *testing.T) {
		claims := Claims{
			StandardClaims: jwt.StandardClaims{
				Subject:  username,
				Issuer:   "qredo",
				Audience: "qredo",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(jwtKey)
		require.NoError(t, err)

		observedErr := service.VerifyToken(context.TODO(), signedToken)

		assert.True(t, errors.Is(observedErr, ErrTokenInvalidExpiration))
	})

	t.Run("invalid token string", func(t *testing.T) {
		observedErr := service.VerifyToken(context.TODO(), "invalid-token-string")
		assert.Error(t, observedErr)
	})
}

func TestDefaultService_Sum(t *testing.T) {
	service := NewDefaultService(zap.NewNop(), nil)

	data := []float64{1, 2, 3}
	expectedHash := "2270ab850480a8ade7647dc3066dde96209bc0314b4847a619e1231e334c00ad"

	actualHash, err := service.Sum(context.TODO(), data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if actualHash != expectedHash {
		t.Errorf("Unexpected hash value: got %v, want %v", actualHash, expectedHash)
	}
}

func TestSumNumbers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		data          any
		expectedSum   float64
		expectedError error
	}{
		{
			name:          "float64",
			data:          1.0,
			expectedSum:   1.0,
			expectedError: nil,
		},
		{
			name:        "int",
			data:        1,
			expectedSum: 1.0,
		},
		{
			name:          "string",
			data:          "1",
			expectedSum:   1.0,
			expectedError: nil,
		},
		{
			name:        "string is empty",
			data:        "",
			expectedSum: 0,
		},
		{
			name:          "string is not a number",
			data:          "a",
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "bool",
			data:          true,
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "slice of float64",
			data:          []float64{1.0, 3.2},
			expectedSum:   4.2,
			expectedError: nil,
		},
		{
			name:          "slice of int",
			data:          []int{1, 2, 3, 4},
			expectedSum:   10,
			expectedError: nil,
		},
		{
			name:          "slice of string",
			data:          []string{"1", "3"},
			expectedSum:   4.0,
			expectedError: nil,
		},
		{
			name:          "slice of bool",
			data:          []bool{true, false},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "slice of mixed types",
			data:          []any{1.0, "3"},
			expectedSum:   4.0,
			expectedError: nil,
		},
		{
			name:          "slice of mixed types with an error",
			data:          []any{1.0, "a"},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:        "slice is empty",
			data:        []any{},
			expectedSum: 0,
		},
		{
			name:        "slice contains an empty string",
			data:        []any{""},
			expectedSum: 0,
		},
		{
			name:        "slice contains an empty slice",
			data:        []any{[]any{}},
			expectedSum: 0,
		},
		{
			name:        "slice contains a slice with another slice",
			data:        []any{[]any{[]any{1.0, 2.0}}},
			expectedSum: 3.0,
		},
		{
			name:          "map",
			data:          map[string]any{"a": 6, "b": 4},
			expectedSum:   10,
			expectedError: nil,
		},
		{
			name:          "map is empty",
			data:          map[string]any{},
			expectedSum:   0,
			expectedError: nil,
		},
		{
			name:          "nil",
			data:          nil,
			expectedSum:   0,
			expectedError: nil,
		},
		{
			name:          "pointer",
			data:          new(int),
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "function",
			data:          func() {},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "channel",
			data:          make(chan int),
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "interface",
			data:          new(any),
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "unsupported type",
			data:          struct{}{},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "slice contains unsupported type",
			data:          []any{1.0, true},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
		{
			name:          "map contains unsupported type",
			data:          map[string]any{"a": 1.0, "b": true},
			expectedSum:   0,
			expectedError: ErrUnsupportedValueType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			observedSum, observedErr := sumNumbers(tc.data)

			assert.Equal(t, tc.expectedSum, observedSum)
			assert.True(t, errors.Is(observedErr, tc.expectedError))
		})
	}
}
