package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/kardianos/osext"
)

// Filestore is a hacky filesystem "database". To be removed.
type Filestore struct {
	folder string
}

// NewFilestore creates a new Datastore database object
func NewFilestore(ctx context.Context, cfg *config.Config) (Database, error) {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return nil, err
	}
	return &Filestore{
		folder: folder,
	}, nil
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

// Remove is used to delete a file from the filesystem
func (db *Filestore) Remove(buildID string) error {
	file := db.makeFileName(buildID)
	return os.Remove(file)
}
