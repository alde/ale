package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alde/ale"

	"github.com/alde/ale/config"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Test_ServiceMetadata(t *testing.T) {
	m := mux.NewRouter()
	config := &config.Config{}

	h := NewHandler(config, mockDatabase)
	m.HandleFunc("/service-metadata", h.ServiceMetadata())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/service-metadata", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, wr.Code, http.StatusOK)

	var actual map[string]interface{}
	err := json.Unmarshal(wr.Body.Bytes(), &actual)
	assert.Nil(t, err)

	expectedKeys := []string{
		"service_name", "service_version", "description", "owner", "database",
	}

	for _, k := range expectedKeys {
		_, ok := actual[k]
		assert.True(t, ok, "expected key %s", k)
	}
}

func Test_ProcessOptions(t *testing.T) {
	m := mux.NewRouter()
	config := &config.Config{}

	h := NewHandler(config, mockDatabase)
	m.HandleFunc("/api/v1/process", h.ProcessOptions())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("OPTIONS", "/api/v1/process", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, wr.Code, http.StatusOK)

	var actual map[string]interface{}
	err := json.Unmarshal(wr.Body.Bytes(), &actual)
	assert.Nil(t, err)
	assert.Equal(t, []string{"*"}, wr.HeaderMap["Access-Control-Allow-Origin"])
	assert.Equal(t, []string{"POST, GET, OPTIONS"}, wr.HeaderMap["Access-Control-Allow-Methods"])
	assert.Equal(t, []string{"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"}, wr.HeaderMap["Access-Control-Allow-Headers"])
}

func Test_GetJenkinsBuild(t *testing.T) {
	m := mux.NewRouter()
	config := &config.Config{}
	jdata := &ale.JenkinsData{
		BuildID: "buildId",
		Status:  "IN_PROGRESS",
	}
	mockDatabase.Put(jdata, "buildId")

	h := NewHandler(config, mockDatabase)
	m.HandleFunc("/api/v1/build/{id}", h.GetJenkinsBuild())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/api/v1/build/buildId", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, http.StatusOK, wr.Code)
	var actual ale.JenkinsData
	body, _ := ioutil.ReadAll(io.LimitReader(wr.Body, 1048576))
	json.Unmarshal(body, &actual)
	assert.Equal(t, jdata, &actual)
}
