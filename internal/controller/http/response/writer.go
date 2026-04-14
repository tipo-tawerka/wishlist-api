package response

import (
	"encoding/json"
	"net/http"
)

type HTTPError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int) {
	err := http.StatusText(statusCode)
	if err == "" {
		err = "Unknown error"
	}
	WriteJSON(w, HTTPError{Error: err}, statusCode)
}
