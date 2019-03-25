package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
	"github.com/alde/ale/mock"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var cfg0 = config.DefaultConfig()

func Test_ServiceMetadata(t *testing.T) {
	m := mux.NewRouter()

	h := NewHandler(cfg0, mockDatabase)
	m.HandleFunc("/service-metadata", h.ServiceMetadata())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/service-metadata", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, wr.Code, http.StatusOK)

	var actual map[string]interface{}
	err := json.Unmarshal(wr.Body.Bytes(), &actual)
	assert.Nil(t, err)

	expectedKeys := []string{
		"service_name", "service_version", "description", "owner", "service_version", "build_date",
	}

	for _, k := range expectedKeys {
		_, ok := actual[k]
		assert.True(t, ok, "expected key %s", k)
	}
}

func Test_ProcessOptions(t *testing.T) {
	m := mux.NewRouter()

	h := NewHandler(cfg0, mockDatabase)
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
	jdata := &ale.JenkinsData{
		BuildID: "buildId",
		Status:  "IN_PROGRESS",
	}
	mockDatabase.Put(jdata, "buildId")

	h := NewHandler(cfg0, mockDatabase)
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

func Test_GetJenkinsBuildNotFound(t *testing.T) {
	m := mux.NewRouter()
	dbclient := &mock.Datastore{}
	database := &db.Datastore{
		Client: dbclient,
	}

	h := NewHandler(cfg0, database)
	m.HandleFunc("/api/v1/build/{id}", h.GetJenkinsBuild())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/api/v1/build/buildId0", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, http.StatusNotFound, wr.Code)
	actual := `{"buildID":"buildId0","message":"build not found in database, has it been processed?"}`
	body, _ := ioutil.ReadAll(io.LimitReader(wr.Body, 1048576))
	assert.Equal(t, strings.Trim(string(body), "\n"), actual)
}

func Test_GetJenkinsBuildError(t *testing.T) {
	m := mux.NewRouter()
	dbclient := &mock.Datastore{
		GetFn: func(context.Context, *datastore.Key, interface{}) error {
			return errors.New("a communication problem with datastore")
		},
		CountFn: func(context.Context, *datastore.Query) (int, error) {
			return 1, nil
		},
	}
	database := &db.Datastore{
		Client: dbclient,
	}

	h := NewHandler(cfg0, database)
	m.HandleFunc("/api/v1/build/{id}", h.GetJenkinsBuild())
	wr := httptest.NewRecorder()

	r, _ := http.NewRequest("GET", "/api/v1/build/buildId9", nil)
	m.ServeHTTP(wr, r)

	assert.Equal(t, http.StatusUnprocessableEntity, wr.Code)
	actual := `"Message":"unable to query from database"`
	body, _ := ioutil.ReadAll(io.LimitReader(wr.Body, 1048576))
	assert.Contains(t, string(body), actual)
}
