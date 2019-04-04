package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	// Import postgres driver into the scope of this package (required)
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
)

// SQL struct implementing the Database interface with PostgreSQL backend
type SQL struct {
	db *sql.DB
}

func readPasswordFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}

	password, _ := ioutil.ReadAll(file)
	return strings.Trim(string(password), "")
}

// New is used to create a new PostgreSQL database connection
func New(cfg *config.Config) (db.Database, error) {
	password := readPasswordFile(cfg.PostgreSQL.PasswordFile)

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Username,
		password,
		cfg.PostgreSQL.Database,
	)

	if cfg.PostgreSQL.DisableSSL {
		connectionString += " sslmode=disable"
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logrus.
			WithField("connectionString", connectionString).
			WithError(err).
			Error("failed to initialize driver")
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logrus.
			WithField("connectionString", connectionString).
			WithError(err).
			Error("failed to ping database")
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS ale_jenkins_logs ( \n" +
		" build_id VARCHAR(128) NOT NULL, \n" +
		" log jsonb NOT NULL, \n" +
		" PRIMARY KEY (build_id) \n" +
		")")
	if err != nil {
		logrus.
			WithError(err).
			Error("unable to create database table")
		return nil, err
	}

	return &SQL{
		db: db,
	}, nil
}

// Put inserts data into the database
func (sql *SQL) Put(data *ale.JenkinsData, buildID string) error {
	j, err := json.Marshal(data)
	if err != nil {
		logrus.
			WithField("buildId", buildID).
			WithError(err).
			Error("error marshalling data into json")
	}
	jstring := string(j)
	query := `INSERT INTO ale_jenkins_logs(build_id, log)
		 VALUES ($1, $2::jsonb)
		 ON CONFLICT (build_id) DO UPDATE SET log = $3::jsonb`
	_, err = sql.db.Exec(query, buildID, jstring, jstring)

	if err != nil {
		logrus.
			WithField("buildId", buildID).
			WithError(err).
			Error("error inserting logs into database")
		return err
	}

	logrus.
		WithField("buildId", buildID).
		Info("inserted log into database")

	return nil
}

// Has verifies the existance of a log
func (sql *SQL) Has(buildID string) (bool, error) {
	// TODO: don't re-use Get
	b, err := sql.Get(buildID)
	if err != nil {
		return false, err
	}
	if (&ale.JenkinsData{}) == b {
		return false, nil
	}
	return true, nil
}

// Get retrieves logs from the database
func (sql *SQL) Get(buildID string) (*ale.JenkinsData, error) {
	query := "SELECT log FROM ale_jenkins_logs WHERE build_id = $1"
	row := sql.db.QueryRow(query, buildID)
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		logrus.
			WithError(err).
			WithFields(logrus.Fields{
				"buildId": buildID,
				"data":    data,
			}).
			Error("unable to deserialize data")
		return nil, err
	}

	var jlog ale.JenkinsData
	err = json.Unmarshal([]byte(data), &jlog)
	return &jlog, err
}

// Remove is used to remove an entry from the database
func (sql *SQL) Remove(buildID string) error {
	query := "DELETE FROM ale_jenkins_logs WHERE build_id = $1"
	_, err := sql.db.Exec(query, buildID)
	if err != nil {
		logrus.
			WithField("buildId", buildID).
			WithError(err).
			Error("error deleting logs from database")
	} else {
		logrus.
			WithField("buildId", buildID).
			Info("deleted log from database")
	}
	return err
}
