package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DefaultConfig(t *testing.T) {
	c := DefaultConfig()
	assert := assert.New(t)
	assert.Equal("0.0.0.0", c.Address)
	assert.Equal(7654, c.Port)
	assert.Equal("text", c.LogFormat)
	assert.Equal("debug", c.LogLevel)
	assert.Equal("file", c.Database.Type)
	assert.Equal(os.Getenv("USER"), c.Owner)

}

func Test_ReadEnvironment(t *testing.T) {
	c := DefaultConfig()
	assert := assert.New(t)

	os.Setenv("ALE_ADDRESS", "10.0.0.0")
	os.Setenv("ALE_PORT", "9090")
	os.Setenv("ALE_LOGLEVEL", "error")
	os.Setenv("ALE_LOGFORMAT", "json")
	os.Setenv("ALE_OWNER", "the_boss")
	os.Setenv("ALE_DATABASE_TYPE", "datastore")
	os.Setenv("ALE_DATABASE_PROJECT", "my-gcp-project")
	os.Setenv("ALE_DATABASE_NAMESPACE", "my-namespace")

	ReadEnvironment(c)

	os.Unsetenv("ALE_ADDRESS")
	os.Unsetenv("ALE_PORT")
	os.Unsetenv("ALE_LOGLEVEL")
	os.Unsetenv("ALE_LOGFORMAT")
	os.Unsetenv("ALE_OWNER")
	os.Unsetenv("ALE_DATABASE_TYPE")
	os.Unsetenv("ALE_DATABASE_PROJECT")
	os.Unsetenv("ALE_DATABASE_NAMESPACE")

	assert.Equal("10.0.0.0", c.Address)
	assert.Equal(9090, c.Port)
	assert.Equal("json", c.LogFormat)
	assert.Equal("error", c.LogLevel)
	assert.Equal("the_boss", c.Owner)
	assert.Equal("datastore", c.Database.Type)
	assert.Equal("my-gcp-project", c.Database.Project)
	assert.Equal("my-namespace", c.Database.Namespace)
}

func Test_ReadConfigFile(t *testing.T) {
	c := DefaultConfig()
	wd, _ := os.Getwd()
	assert := assert.New(t)

	ReadConfigFile(c, fmt.Sprintf("%s/config_test.yml", wd))
	assert.Equal("127.0.0.1", c.Address)
	assert.Equal(8080, c.Port)
	assert.Equal("json", c.LogFormat)
	assert.Equal("info", c.LogLevel)
	assert.Equal("the_team", c.Owner)
	assert.Equal("datastore", c.Database.Type)
	assert.Equal("my-gcs-project", c.Database.Project)
	assert.Equal("ale-jenkinslog", c.Database.Namespace)
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
