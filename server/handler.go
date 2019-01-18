package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/alde/ale/config"
	"github.com/alde/ale/version"
	"github.com/google/uuid"
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

// ProcessRequest represents the json payload of the request
type ProcessRequest struct {
	Org      string `json:"org"`
	Repo     string `json:"repo"`
	BuildID  string `json:"buildId,omitempty"`
	BuildURL string `json:"buildUrl"`
}

type ProcessResponse struct {
	Location string `json:"location"`
}

// ProcessBuild Triggers of a job to process a given build
func (h *Handler) ProcessBuild(conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ProcessRequest
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			handleError(err, w, "unable to read payload")
			return
		}
		err = json.Unmarshal(body, &request)
		if err != nil {
			handleError(err, w, "unable to deserialize payload")
		}
		if request.BuildID == "" {
			request.BuildID = uuid.New().String()
		}

		url := absURL(r, fmt.Sprintf("/api/v1/build/%s", request.BuildID), conf)
		response := &ProcessResponse{
			Location: url,
		}
		writeJSON(http.StatusCreated, response, w)
		return
	}
}

// GetJenkinsBuild returns data about the given build
func (h *Handler) GetJenkinsBuild() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(http.StatusNotImplemented, "Method not implemented yet", w)
	}
}
