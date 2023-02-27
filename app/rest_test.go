package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/qredo-external/go-alessandro-resta/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRESTApp(t *testing.T) {
	givenRouter := chi.NewRouter()
	givenService := &service.DefaultService{}
	givenLogger := zap.NewExample()
	givenPort := "8080"

	app := NewRESTApp(givenLogger, givenPort, givenRouter, givenService)

	assert.NotNil(t, app)
	assert.NotNil(t, app.logger)
	assert.Equal(t, net.JoinHostPort("", givenPort), app.httpServer.Addr)
	assert.NotNil(t, app.httpServer.Handler)
}

func TestAuthHandler(t *testing.T) {
	var genTokenFuncWasCalled bool
	mockSvc := &service.MockService{
		GenerateTokenFunc: func(ctx context.Context, creds service.Credentials) (*service.Token, error) {
			genTokenFuncWasCalled = true
			return &service.Token{
				AccessToken: "foo-token",
				TokenType:   "bar-token-type",
				ExpiresIn:   3600,
			}, nil
		},
	}

	router := chi.NewRouter()

	app := &RESTApp{
		logger: zap.NewNop(),
		svc:    mockSvc,
	}

	router.Post("/auth", app.authHandler)

	body := `{"username": "test-user", "password": "test-pass"}`

	req, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(body))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.True(t, genTokenFuncWasCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t,
		`{"access_token":"foo-token","token_type":"bar-token-type","expired_in":3600}`,
		strings.TrimSpace(w.Body.String()),
	)
}

func TestAuthHandler_InvalidRequest(t *testing.T) {
	var genTokenFuncWasCalled bool
	mockSvc := &service.MockService{
		GenerateTokenFunc: func(ctx context.Context, creds service.Credentials) (*service.Token, error) {
			genTokenFuncWasCalled = true
			return &service.Token{
				AccessToken: "foo-token",
				TokenType:   "bar-token-type",
				ExpiresIn:   3600,
			}, nil
		},
	}

	router := chi.NewRouter()

	app := &RESTApp{
		logger: zap.NewNop(),
		svc:    mockSvc,
	}

	router.Post("/auth", app.authHandler)

	body := `{"username": "test-user"}` // missing password

	req, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(body))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.False(t, genTokenFuncWasCalled)
	assert.Equal(t, ErrInvalidPassword.StatusCode, w.Code)

	var respErr APIError
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&respErr))

	assert.Equal(t, ErrInvalidPassword, respErr)
}

func TestAuthHandler_serviceError(t *testing.T) {
	mockSvc := &service.MockService{
		GenerateTokenFunc: func(ctx context.Context, creds service.Credentials) (*service.Token, error) {
			return nil, errors.New("foo-error")
		},
	}

	router := chi.NewRouter()

	app := &RESTApp{
		logger: zap.NewNop(),
		svc:    mockSvc,
	}

	router.Post("/auth", app.authHandler)

	body := `{"username": "test-user", "password": "test-pass"}`

	req, err := http.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(body))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var respErr APIError
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&respErr))

	assert.Equal(t, ErrInternal, respErr)
}

func TestSumHandler(t *testing.T) {
	mockSvc := &service.MockService{
		VerifyTokenFunc: func(ctx context.Context, token string) error {
			return nil
		},
		SumFunc: func(ctx context.Context, data any) (string, error) {
			return "abcd", nil
		},
	}

	router := chi.NewRouter()

	app := &RESTApp{
		logger: zap.NewNop(),
		svc:    mockSvc,
	}

	router.Post("/sum", app.sumHandler)

	body := `{"a": 2, "b": 3}`

	req, err := http.NewRequest(http.MethodPost, "/sum", bytes.NewBufferString(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer abcd") // token is required

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"sum":"abcd"}`, strings.TrimSpace(w.Body.String()))
}

func TestSumHandler_serviceError(t *testing.T) {
	mockSvc := &service.MockService{
		VerifyTokenFunc: func(ctx context.Context, token string) error {
			return nil
		},
		SumFunc: func(ctx context.Context, data any) (string, error) {
			return "", errors.New("foo-error")
		},
	}

	router := chi.NewRouter()

	app := &RESTApp{
		logger: zap.NewNop(),
		svc:    mockSvc,
	}

	router.Post("/sum", app.sumHandler)

	body := `{"a": 2, "b": 3}`

	req, err := http.NewRequest(http.MethodPost, "/sum", bytes.NewBufferString(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer abcd")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var respErr APIError
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&respErr))

	assert.Equal(t, ErrInternal, respErr)
}

func TestSumHandler_invalidToken(t *testing.T) {
	router := chi.NewRouter()

	app := &RESTApp{logger: zap.NewNop()}

	router.Post("/sum", app.sumHandler)

	body := `{"a": 2, "b": 3}`

	req, err := http.NewRequest(http.MethodPost, "/sum", bytes.NewBufferString(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var respErr APIError
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&respErr))

	assert.Equal(t, ErrUnauthorized, respErr)
}
