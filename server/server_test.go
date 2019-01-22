package server

import (
	"testing"

	"github.com/alde/ale"

	"github.com/alde/ale/config"

	"github.com/stretchr/testify/assert"
)

var (
	cfg          = &config.Config{}
	mockDatabase = &mockDB{
		memory: make(map[string]*ale.JenkinsData),
	}
)

type mockDB struct {
	memory map[string]*ale.JenkinsData
}

// Put inserts data into the database
func (db *mockDB) Put(data *ale.JenkinsData, buildID string) error {
	db.memory[buildID] = data
	return nil
}

// Get retrieves data from the database
func (db *mockDB) Get(buildID string) (*ale.JenkinsData, error) {
	return db.memory[buildID], nil
}

func Test_NewRouter(t *testing.T) {
	h := NewHandler(cfg, mockDatabase)
	nr := NewRouter(cfg, mockDatabase)

	for _, r := range routes(h) {
		assert.NotNil(t, nr.GetRoute(r.Name))
	}
}

func Test_routes(t *testing.T) {
	h := NewHandler(cfg, mockDatabase)
	assert.Len(t, routes(h), 3, "3 routes is the magic number.")
}
