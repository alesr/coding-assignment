package app

import (
	"encoding/json"
	"net/http"
)

func writeJSONError(w http.ResponseWriter, e error) {
	w.Header().Set("Content-Type", "application/json")

	if apiError, ok := e.(APIError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiError.StatusCode)
		json.NewEncoder(w).Encode(apiError)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(ErrInternal)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
