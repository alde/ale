package db

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/alde/ale"

	"github.com/alde/ale/config"
	"github.com/alde/ale/mock"
	"github.com/stretchr/testify/assert"
)

var (
	cfg = &config.Config{
		GoogleCloudDatastore: config.DatastoreConf{
			Namespace: "test-namespace",
			Project:   "gcp-project-1",
		},
	}
	ctx = context.Background()
)

func Test_CreateDatastore(t *testing.T) {
	database, err := NewDatastore(ctx, cfg, &mock.Datastore{})
	assert.NotNil(t, database)
	assert.Nil(t, err)
}

func Test_makeKey(t *testing.T) {
	database, _ := NewDatastore(ctx, cfg, &mock.Datastore{})
	key := database.(*Datastore).makeKey("foobar")

	assert.Equal(t, "JenkinsBuild", key.Kind)
	assert.Equal(t, "foobar", key.Name)
	assert.Equal(t, cfg.GoogleCloudDatastore.Namespace, key.Namespace)
	assert.Nil(t, key.Parent)
}

func Test_Put(t *testing.T) {
	m := &mock.Datastore{}
	database := &Datastore{
		Client: m,
	}
	database.Put(&ale.JenkinsData{}, "foobar")
	assert.True(t, m.PutFnInvoked)
}

func Test_Get(t *testing.T) {
	m := &mock.Datastore{}
	database := &Datastore{
		Client: m,
	}
	database.Put(&ale.JenkinsData{}, "foobar")

	j, _ := database.Get("foobar")
	assert.True(t, m.GetFnInvoked)
	assert.NotNil(t, j)
}

func Test_GetWithError(t *testing.T) {
	m := &mock.Datastore{
		GetFn: func(context.Context, *datastore.Key, interface{}) error {
			return errors.New("a communication problem with datastore")
		},
	}
	database := &Datastore{
		Client: m,
	}

	database.Put(&ale.JenkinsData{}, "foobar")

	_, err := database.Get("foobar")
	assert.True(t, m.GetFnInvoked)
	assert.NotNil(t, err)
}

func Test_Has(t *testing.T) {
	m := &mock.Datastore{}
	database := &Datastore{
		Client: m,
	}
	b, _ := database.Has("foobar")
	assert.True(t, m.CountFnInvoked)
	assert.False(t, b)

	database.Put(&ale.JenkinsData{}, "foobar")

	b, _ = database.Has("foobar")
	assert.True(t, m.CountFnInvoked)
	assert.True(t, b)
}

func Test_HasNotButNoError(t *testing.T) {
	m := &mock.Datastore{
		CountFn: func(context.Context, *datastore.Query) (int, error) {
			return 0, nil
		},
	}
	database := &Datastore{
		Client: m,
	}
	b, _ := database.Has("foobar")
	assert.True(t, m.CountFnInvoked)
	assert.False(t, b)
}
