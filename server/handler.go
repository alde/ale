package server

import (
	"net/http"

	"github.com/alde/ale/config"
	"github.com/alde/ale/version"
)

// Handler holds the server context
type Handler struct {
	config *config.Config
}

// NewHandler createss a new HTTP handler
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{config: cfg}
}

// ServiceMetadata displays hopefully useful information about the service
func (h *Handler) ServiceMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})

		data["owner"] = h.config.Owner
		data["description"] = "Jenkins Build Information"
		data["service_name"] = "ale"
		data["service_version"] = version.Version
		data["gcsbucket"] = h.config.Bucket

		writeJSON(http.StatusOK, data, w)
	}
}

// ProcessBuild Triggers of a job to process a given build
func (h *Handler) ProcessBuild() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(http.StatusNotImplemented, "Method not implemented yet", w)
	}
}

// GetJenkinsBuild returns data about the given build
func (h *Handler) GetJenkinsBuild() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(http.StatusNotImplemented, "Method not implemented yet", w)
	}
}
