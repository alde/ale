package mock

import "github.com/alde/ale"

// DB holds a mocked in-memory database
type DB struct {
	Memory map[string]*ale.JenkinsData
}

// Put inserts data into the database
func (db *DB) Put(data *ale.JenkinsData, buildID string) error {
	db.Memory[buildID] = data
	return nil
}

// Get retrieves data from the database
func (db *DB) Get(buildID string) (*ale.JenkinsData, error) {
	return db.Memory[buildID], nil
}

// Has checks the existance in the database
func (db *DB) Has(buildID string) (bool, error) {
	_, ok := db.Memory[buildID]
	return ok, nil
}
