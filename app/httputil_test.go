package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSONError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		givenError       error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "given APIError",
			givenError:       ErrInvalidPassword,
			expectedStatus:   ErrInvalidPassword.StatusCode,
			expectedResponse: ErrInvalidPassword.Error(),
		},
		{
			name:             "given error",
			givenError:       fmt.Errorf("foo"),
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: ErrInternal.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			writeJSONError(w, tc.givenError)

			require.Equal(t, tc.expectedStatus, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var observedErr APIError
			require.NoError(t, json.NewDecoder(w.Body).Decode(&observedErr))

			assert.Equal(t, tc.expectedResponse, observedErr.Error())
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		givenData    any
		expectedJSON string
	}{
		{
			name:         "given data is nil",
			givenData:    nil,
			expectedJSON: "null",
		},
		{
			name:         "given data is empty",
			givenData:    map[string]string{},
			expectedJSON: "{}",
		},
		{
			name:         "given data is not empty",
			givenData:    map[string]string{"foo": "bar"},
			expectedJSON: `{"foo":"bar"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			writeJSON(w, tc.givenData)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))

			assert.Equal(t, tc.expectedJSON, strings.TrimSpace(w.Body.String()))
		})
	}
}
