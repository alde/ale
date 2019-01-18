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
	assert.Equal(c.Address, "0.0.0.0")
	assert.Equal(c.Port, 7654)
	assert.Equal(c.LogFormat, "text")
	assert.Equal(c.LogLevel, "debug")
	assert.Equal(c.Bucket, "")
	assert.Equal(c.Owner, os.Getenv("USER"))

}

func Test_ReadEnvironment(t *testing.T) {
	c := DefaultConfig()
	assert := assert.New(t)

	os.Setenv("ALE_ADDRESS", "10.0.0.0")
	os.Setenv("ALE_PORT", "9090")
	os.Setenv("ALE_LOGLEVEL", "error")
	os.Setenv("ALE_LOGFORMAT", "json")
	os.Setenv("ALE_GCSBUCKET", "testbucket")
	os.Setenv("ALE_OWNER", "the_boss")

	ReadEnvironment(c)

	os.Unsetenv("ALE_ADDRESS")
	os.Unsetenv("ALE_PORT")
	os.Unsetenv("ALE_LOGLEVEL")
	os.Unsetenv("ALE_LOGFORMAT")
	os.Unsetenv("ALE_GCSBUCKET")
	os.Unsetenv("ALE_OWNER")

	assert.Equal(c.Address, "10.0.0.0")
	assert.Equal(c.Port, 9090)
	assert.Equal(c.LogFormat, "json")
	assert.Equal(c.LogLevel, "error")
	assert.Equal(c.Owner, "the_boss")
	assert.Equal(c.Bucket, "testbucket")
}

func Test_ReadConfigFile(t *testing.T) {
	c := DefaultConfig()
	wd, _ := os.Getwd()
	assert := assert.New(t)

	ReadConfigFile(c, fmt.Sprintf("%s/config_test.yml", wd))
	assert.Equal(c.Address, "127.0.0.1")
	assert.Equal(c.Port, 8080)
	assert.Equal(c.LogFormat, "json")
	assert.Equal(c.LogLevel, "info")
	assert.Equal(c.Owner, "the_team")
	assert.Equal(c.Bucket, "testbucket")
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
