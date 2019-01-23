package db

import (
	"context"
	"testing"

	"github.com/alde/ale/config"
	"github.com/stretchr/testify/assert"
)

var (
	cfg = &config.Config{
		Database: config.DBConfig{
			Type:      "datastore",
			Namespace: "test-namespace",
			Project:   "gcp-project-1",
		},
	}
	ctx = context.Background()
)

func Test_CreateDatastore(t *testing.T) {
	database, err := NewDatastore(ctx, cfg)
	assert.NotNil(t, database)
	assert.Nil(t, err)
}

func Test_FailCreatingDatastore(t *testing.T) {
	database, err := NewDatastore(ctx, &config.Config{})
	assert.Nil(t, database)
	assert.NotNil(t, err)
}
