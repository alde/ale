package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/datastore"
	"github.com/Sirupsen/logrus"
	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/kardianos/osext"
)

// Database interface providing the contract that we expect
type Database interface {
	// Put inserts data into the database
	Put(data *ale.JenkinsData, buildID string) error
	Get(buildID string) (*ale.JenkinsData, error)
}

// Datastore is a Google Cloud Datastore implementation of the Database interface
type Datastore struct {
	client    *datastore.Client
	ctx       context.Context
	namespace string
}

// NewDatastore creates a new Datastore database object
func NewDatastore(ctx context.Context, cfg *config.Config) Database {
	dsClient, err := datastore.NewClient(ctx, cfg.Database.Project)
	if err != nil {
		logrus.WithError(err).Error("unable to create datastore client")
	}
	return &Datastore{
		client:    dsClient,
		ctx:       ctx,
		namespace: cfg.Database.Namespace,
	}
}

// Put inserts data into the database
func (db *Datastore) Put(data *ale.JenkinsData, buildID string) error {
	key := &datastore.Key{
		Kind:      "JenkinsBuild",
		Name:      buildID,
		Parent:    nil,
		Namespace: db.namespace,
	}
	_, err := db.client.Put(db.ctx, key, data)
	return err
}

// Get retrieves data from the database
func (db *Datastore) Get(buildID string) (*ale.JenkinsData, error) {
	var entity ale.JenkinsData
	key := &datastore.Key{
		Kind:      "JenkinsBuild",
		Name:      buildID,
		Parent:    nil,
		Namespace: db.namespace,
	}
	err := db.client.Get(db.ctx, key, entity)
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

// Filestore is a hacky filesystem "database". To be removed.
type Filestore struct {
	folder string
}

// NewFilestore creates a new Datastore database object
func NewFilestore(ctx context.Context, cfg *config.Config) Database {
	folder, _ := osext.ExecutableFolder()
	return &Filestore{
		folder: folder,
	}
}

// Put writes a file to the filesystem
func (db *Filestore) Put(data *ale.JenkinsData, buildID string) error {
	b, _ := json.MarshalIndent(data, "", "\t")
	file := fmt.Sprintf("%s/out_%s.json", db.folder, buildID)
	err := ioutil.WriteFile(file, b, 0644)
	if err != nil {
		logrus.WithError(err).Error("error writing file")
		return err
	}
	logrus.WithField("file", file).Debug("file written")
	return nil
}

// Get reads a file from the filesystem
func (db *Filestore) Get(buildID string) (*ale.JenkinsData, error) {
	file := fmt.Sprintf("%s/out_%s.json", db.folder, buildID)
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var resp ale.JenkinsData
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
