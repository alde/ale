package db

import (
	"github.com/alde/ale"
)

// Database interface providing the contract that we expect
type Database interface {
	Put(data *ale.JenkinsData, buildID string) error
	Get(buildID string) (*ale.JenkinsData, error)
	Has(buildID string) (bool, error)
}
