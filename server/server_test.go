package server

import (
	"testing"

	"github.com/alde/ale/config"

	"github.com/stretchr/testify/assert"
)

var (
	cfg = &config.Config{}
)

func Test_NewRouter(t *testing.T) {
	h := NewHandler(cfg)
	nr := NewRouter(cfg)

	for _, r := range routes(h) {
		assert.NotNil(t, nr.GetRoute(r.Name))
	}
}

func Test_routes(t *testing.T) {
	h := NewHandler(cfg)
	assert.Len(t, routes(h), 3, "3 routes is the magic number.")
}
