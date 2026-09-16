package utils

import (
	"encoding/json"
	"net/http"
)

type ResponseEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ResponseEnvelope{
		Success: status >= 200 && status < 300,
		Message: message,
		Data:    data,
	})
}

func WriteError(w http.ResponseWriter, status int, errMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ResponseEnvelope{
		Success: false,
		Error:   errMessage,
	})
}