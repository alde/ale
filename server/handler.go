package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/alde/ale/db"

	"github.com/alde/ale/config"
	"github.com/alde/ale/jenkins"
	"github.com/alde/ale/version"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler holds the server context
type Handler struct {
	config         *config.Config
	database       db.Database
	crawlerCreator func(db.Database, *config.Config) *jenkins.Crawler
}

// NewHandler createss a new HTTP handler
func NewHandler(cfg *config.Config, db db.Database) *Handler {
	return &Handler{config: cfg, database: db, crawlerCreator: jenkins.NewCrawler}
}

// ServiceMetadata displays hopefully useful information about the service
func (h *Handler) ServiceMetadata() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})

		data["owner"] = h.config.Owner
		data["description"] = "Jenkins Build Information"
		data["service_name"] = "ale"
		data["service_version"] = version.Version
		data["database"] = h.config.Database.Type

		writeJSON(http.StatusOK, data, w)
	}
}

// ProcessRequest represents the json payload of the request
type ProcessRequest struct {
	BuildID  string `json:"buildId,omitempty"`
	BuildURL string `json:"buildUrl"`
	Recrawl  bool   `json:"forceRecrawl,omitempty"`
}

// ProcessResponse represents the response from a requested processing
type ProcessResponse struct {
	Location string `json:"location"`
}

// ProcessBuild Triggers of a job to process a given build
func (h *Handler) ProcessBuild() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		request := &ProcessRequest{
			Recrawl: false,
		}
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			handleError(err, w, "unable to read payload")
			return
		}
		err = json.Unmarshal(body, &request)
		if err != nil {
			handleError(err, w, "unable to deserialize payload")
			return
		}
		if request.BuildURL == "" {
			handleError(err, w, "build_url is required")
			return
		}
		if resp0, err := http.Head(request.BuildURL); err != nil || resp0.StatusCode != http.StatusOK {
			writeJSON(http.StatusBadRequest, "error checking build_url", w)
			return
		}
		if request.BuildID == "" {
			request.BuildID = uuid.New().String()
		}

		exists, err := h.database.Has(request.BuildID)
		if err != nil {
			logrus.WithError(err).Warn("unable to check for existance of database entry")
		}

		url := absURL(r, fmt.Sprintf("/api/v1/build/%s", request.BuildID), h.config)
		response := &ProcessResponse{
			Location: url,
		}
		if exists && !request.Recrawl {
			writeJSON(http.StatusFound, response, w)
			return
		}
		if exists && request.Recrawl {
			h.database.Remove(request.BuildID)
		}

		go func() {
			crawler := h.crawlerCreator(h.database, h.config)
			crawler.CrawlJenkins(request.BuildURL, request.BuildID)
		}()
		writeJSON(http.StatusCreated, response, w)
		return
	}
}

// GetJenkinsBuild returns data about the given build
func (h *Handler) GetJenkinsBuild() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		vars := mux.Vars(r)
		buildID := vars["id"]
		if exists, _ := h.database.Has(buildID); !exists {
			data := make(map[string]string)
			data["buildID"] = buildID
			data["message"] = "build not found in database, has it been processed?"
			writeJSON(http.StatusNotFound, data, w)
			return
		}
		data, err := h.database.Get(buildID)

		if err != nil {
			handleError(err, w, "unable to query from database")
			return
		}
		writeJSON(http.StatusOK, data, w)
	}
}

// ProcessOptions handles the OPTIONS call for CORS
func (h *Handler) ProcessOptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		writeJSON(http.StatusOK, struct{}{}, w)
	}
}
