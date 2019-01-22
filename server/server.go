package server

import (
	"net/http"

	"github.com/alde/ale/db"

	"github.com/alde/ale/config"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// NewRouter is used to create a new HTTP router
func NewRouter(cfg *config.Config, db db.Database) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	h := NewHandler(cfg, db)

	for _, route := range routes(h) {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(prometheus.InstrumentHandler(route.Name, route.Handler))
	}
	return router
}

// Route enforces the structure of a route
type route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

func routes(h *Handler) []route {
	return []route{
		{
			Name:    "PostBuild",
			Method:  "POST",
			Pattern: "/api/v1/process",
			Handler: h.ProcessBuild(),
		},
		{
			Name:    "GetBuild",
			Method:  "GET",
			Pattern: "/api/v1/build/{id}",
			Handler: h.GetJenkinsBuild(),
		},
		{
			Name:    "ServiceMetadata",
			Method:  "GET",
			Pattern: "/service-metadata",
			Handler: h.ServiceMetadata(),
		},
	}
}
