package api

import (
	"encoding/json"
	"net/http"
)

// MessageToJSON constructs the JSON response from the provided parameters
func MessageToJSON(status int, msg string, content any) JSONResponse {
	return JSONResponse{
		ResponseContent: content,
		HTTPStatus:      status,
		Message:         msg,
	}
}

// WriteJSON writes the JSON response to the http.ResponseWriter
func WriteJSON(w http.ResponseWriter, content JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(content.HTTPStatus)

	// Convert the content to JSON and write to the response
	if err := json.NewEncoder(w).Encode(content); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
