package server

import (
	"testing"

	"github.com/alde/ale"

	"github.com/alde/ale/config"
	"github.com/alde/ale/mock"

	"github.com/stretchr/testify/assert"
)

var (
	cfg          = &config.Config{}
	mockDatabase = &mock.DB{
		Memory: make(map[string]*ale.JenkinsData),
	}
)

func Test_NewRouter(t *testing.T) {
	h := NewHandler(cfg, mockDatabase)
	nr := NewRouter(cfg, mockDatabase)

	for _, r := range routes(h) {
		assert.NotNil(t, nr.GetRoute(r.Name))
	}
}

func Test_routes(t *testing.T) {
	h := NewHandler(cfg, mockDatabase)
	assert.Len(t, routes(h), 4, "4 routes is the magic number.")
}
