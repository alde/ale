package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/Sirupsen/logrus"
	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/kardianos/osext"
)

// Database interface providing the contract that we expect
type Database interface {
	Put(data *ale.JenkinsData, buildID string) error
	Get(buildID string) (*ale.JenkinsData, error)
	Has(buildID string) (bool, error)
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
		logrus.WithError(err).Fatal("unable to create datastore client")
	}
	return &Datastore{
		client:    dsClient,
		ctx:       ctx,
		namespace: cfg.Database.Namespace,
	}
}

// DatastoreEntity is used to store data in datastore, and prevent indexing of the huge json
type DatastoreEntity struct {
	Key   string          `json:"key" datastore:"key"`
	Value ale.JenkinsData `json:"value" datastore:"value,noindex"`
}

func (db *Datastore) makeKey(buildID string) *datastore.Key {
	return &datastore.Key{
		Kind:      "JenkinsBuild",
		Name:      buildID,
		Parent:    nil,
		Namespace: db.namespace,
	}
}

// Put inserts data into the database
func (db *Datastore) Put(data *ale.JenkinsData, buildID string) error {
	key := db.makeKey(buildID)
	entity := &DatastoreEntity{
		Key:   buildID,
		Value: *data,
	}
	_, err := db.client.Put(db.ctx, key, entity)
	return err
}

// Has verifies the existance of a key
func (db *Datastore) Has(buildID string) (bool, error) {
	key := db.makeKey(buildID)
	query := datastore.
		NewQuery("JenkinsBuild").
		Namespace(db.namespace).
		Filter("__key__ =", key).
		Limit(1) // Key should be unique, so limit to 1
	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Debug("checking the existance of database entry")
	count, err := db.client.Count(db.ctx, query)
	if err != nil {
		logrus.WithError(err).WithField("build_id", buildID).Debug("database entry not found")
		return false, err
	}
	if count == 1 {
		logrus.WithFields(logrus.Fields{
			"build_id": buildID,
			"count":    count,
		}).Debug("database entry found")
		return true, err
	}
	logrus.WithFields(logrus.Fields{
		"build_id": buildID,
	}).Debug("database entry not found")
	return false, err
}

// Get retrieves data from the database
func (db *Datastore) Get(buildID string) (*ale.JenkinsData, error) {
	var entity DatastoreEntity
	key := db.makeKey(buildID)
	err := db.client.Get(db.ctx, key, &entity)
	if err != nil {
		return nil, err
	}

	jdata := entity.Value

	return &jdata, nil
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

func (db *Filestore) makeFileName(buildID string) string {
	return fmt.Sprintf("%s/out_%s.json", db.folder, buildID)
}

// Put writes a file to the filesystem
func (db *Filestore) Put(data *ale.JenkinsData, buildID string) error {
	file := db.makeFileName(buildID)
	b, _ := json.MarshalIndent(data, "", "\t")
	err := ioutil.WriteFile(file, b, 0644)
	if err != nil {
		logrus.WithError(err).Error("error writing file")
		return err
	}
	logrus.WithField("file", file).Debug("file written")
	return nil
}

// Has checks for the existance of the file
func (db *Filestore) Has(buildID string) (bool, error) {
	file := db.makeFileName(buildID)
	_, err := os.Open(file)
	if err != nil {
		return false, err
	}
	return true, nil

}

// Get reads a file from the filesystem
func (db *Filestore) Get(buildID string) (*ale.JenkinsData, error) {
	file := db.makeFileName(buildID)
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
