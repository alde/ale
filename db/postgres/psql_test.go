package postgres

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/alde/ale"
	"github.com/alde/ale/config"
)

func setupPostgreSQLTestContainer() *config.Config {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "ale_db_password",
			"POSTGRES_DB":       "ale_db_testing",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	psqlC, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	time.Sleep(30 * time.Second)

	cfg := config.DefaultConfig()
	p, _ := psqlC.MappedPort(ctx, "5432/tcp")
	tmpdir := os.TempDir()
	file, _ := ioutil.TempFile(tmpdir, "sql")
	_, _ = file.WriteString("ale_db_password")

	cfg.PostgreSQL.Host, _ = psqlC.Host(ctx)
	cfg.PostgreSQL.Port = p.Int()
	cfg.PostgreSQL.Username = "postgres"
	cfg.PostgreSQL.PasswordFile = file.Name()
	cfg.PostgreSQL.Database = "ale_db_testing"
	cfg.PostgreSQL.DisableSSL = true

	return cfg
}

func Test_Postgres(t *testing.T) {
	cfg := setupPostgreSQLTestContainer()
	sql, err := New(cfg)

	assert.Nil(t, err)

	t.Run("test inserting into the database", func(t *testing.T) {
		err := sql.Put(&ale.JenkinsData{BuildID: "test_put"}, "test_put")
		assert.Nil(t, err)
	})

	t.Run("test retrieve from db", func(t *testing.T) {
		sql.Put(&ale.JenkinsData{BuildID: "test_get"}, "test_get")
		j, err := sql.Get("test_get")
		assert.Equal(t, "test_get", j.BuildID)
		assert.Nil(t, err)
	})

	t.Run("test inserting replaces in the database", func(t *testing.T) {
		_ = sql.Put(&ale.JenkinsData{BuildID: "test_replace", Name: "Original"}, "test_replace")
		j0, _ := sql.Get("test_replace")
		assert.Equal(t, "Original", j0.Name)
		_ = sql.Put(&ale.JenkinsData{BuildID: "test_replace", Name: "Replacement"}, "test_replace")
		j1, _ := sql.Get("test_replace")
		assert.Equal(t, "Replacement", j1.Name)
	})

	t.Run("test to ensure Get will exit correctly if item isn't found", func(t *testing.T) {
		_, err := sql.Get("test_get_not_found")
		assert.NotNil(t, err)
	})

	t.Run("test to ensure Has returns true if it is in the database", func(t *testing.T) {
		b, _ := sql.Has("test_has")
		assert.False(t, b)

		sql.Put(&ale.JenkinsData{BuildID: "test_has"}, "test_has")
		b, _ = sql.Has("test_has")
		assert.True(t, b)
	})

	t.Run("test to ensure Has returns false if it's not in the database", func(t *testing.T) {
		b, _ := sql.Has("test_has_not")
		assert.False(t, b)
	})

}
