package server

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/alde/ale/config"
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

func handleError(err error, w http.ResponseWriter, message string) {
	if err == nil {
		return
	}

	errorMessage := struct {
		Error   error
		Message string
	}{
		err, message,
	}

	if err = writeJSON(422, errorMessage, w); err != nil {
		logrus.WithError(err).WithField("message", message).Panic("Unable to respond")
	}
}

func absURL(r *http.Request, path string, conf *config.Config) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}

	// if conf.URLPrefix != "" {
	// 	path = fmt.Sprintf("%s%s", conf.URLPrefix, path)
	// 	logrus.WithField("path", path).Debug("absurl was computed")
	// }

	url := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   path,
	}
	return url.String()
}
