package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alde/ale/config"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Test_ServiceMetadata(t *testing.T) {
	m := mux.NewRouter()
	config := &config.Config{}

	h := NewHandler(config)
	m.HandleFunc("/service-metadata", h.ServiceMetadata())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/service-metadata", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, wr.Code, http.StatusOK)

	var actual map[string]interface{}
	err := json.Unmarshal(wr.Body.Bytes(), &actual)
	assert.Nil(t, err)

	expectedKeys := []string{
		"service_name", "service_version", "description", "owner", "gcsbucket",
	}

	for _, k := range expectedKeys {
		_, ok := actual[k]
		assert.True(t, ok, "expected key %s", k)
	}
}
