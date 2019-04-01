package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DefaultConfig(t *testing.T) {
	c := DefaultConfig()
	assert.Equal(t, "0.0.0.0", c.Server.Address)
	assert.Equal(t, 7654, c.Server.Port)
	assert.Equal(t, "text", c.Logging.Format)
	assert.Equal(t, "DEBUG", c.Logging.Level)
	assert.Equal(t, DatastoreConf{}, c.GoogleCloudDatastore)
	assert.Equal(t, os.Getenv("USER"), c.Metadata["owner"])

}

func Test_ReadConfigFile(t *testing.T) {
	c := DefaultConfig()
	wd, _ := os.Getwd()

	ReadConfigFile(c, fmt.Sprintf("%s/config_test.toml", wd))
	assert.Equal(t, "127.0.0.1", c.Server.Address)
	assert.Equal(t, 8080, c.Server.Port)
	assert.Equal(t, "json", c.Logging.Format)
	assert.Equal(t, "INFO", c.Logging.Level)
	assert.Equal(t, "the_team", c.Metadata["owner"])

	assert.NotEqual(t, DatastoreConf{}, c.GoogleCloudDatastore)
	assert.Equal(t, "my-gcs-project", c.GoogleCloudDatastore.Project)
	assert.Equal(t, "ale-jenkinslog", c.GoogleCloudDatastore.Namespace)
}

func Test_ReadConfigFile_Error(t *testing.T) {
	c := DefaultConfig()
	d := DefaultConfig()

	ReadConfigFile(c, getConfigFilePath(""))

	assert.Equal(t, c, d)
}

func Test_getConfigFilePath(t *testing.T) {
	fp := getConfigFilePath("")
	assert.Empty(t, fp)
}

func Test_Initialize(t *testing.T) {
	c := Initialize("")
	d := DefaultConfig()

	assert.Equal(t, c, d)
}
