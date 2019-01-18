package server

import (
	"encoding/json"
	"net/http"
)

const (
	contentTypeJSON = "application/json; charset=UTF-8"
)

func writeJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func notFound(w http.ResponseWriter) error {
	return writeError(http.StatusNotFound, "Not Found", w)
}

func writeError(status int, message string, w http.ResponseWriter) error {
	data := make(map[string]string)
	data["error"] = message
	return writeJSON(status, data, w)
}
