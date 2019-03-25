package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"
)

// DatastoreConf holds the config values for MySQL
type DatastoreConf struct {
	Namespace string
	Project   string
}

// Config struct holds the current configuration
type Config struct {
	Server struct {
		Address string
		Port    int
	}

	Logging struct {
		Format string
		Level  string
	}

	Metadata map[string]string

	GoogleCloudDatastore DatastoreConf

	Crawler struct {
		LogPattern string
	}
}

// Initialize a new Config
func Initialize(configFile string) *Config {
	cfg := DefaultConfig()
	ReadConfigFile(cfg, getConfigFilePath(configFile))

	return cfg
}

// DefaultConfig returns a Config struct with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	cfg.Server.Address = "0.0.0.0"
	cfg.Server.Port = 7654

	cfg.Logging.Format = "text"
	cfg.Logging.Level = "DEBUG"

	cfg.Metadata = make(map[string]string)
	cfg.Metadata["owner"] = os.Getenv("USER")

	cfg.Crawler.LogPattern = `.*\[([\d{4}\-\d{2}\-\d{2}T\d{2}:\d{2}:\d{2}.\d*Z]*)\].*?\s(.*)$`

	return cfg
}

// getConfigFilePath returns the location of the config file in order of priority:
// 1 ) --config commandline flag
// 1 ) File in same directory as the executable
// 2 ) Global file in /etc/ale/config.toml
func getConfigFilePath(configPath string) string {
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		panic(fmt.Sprintf("Unable to open %s.", configPath))
	}
	path, _ := osext.ExecutableFolder()
	path = fmt.Sprintf("%s/config.toml", path)
	if _, err := os.Open(path); err == nil {
		return path
	}
	globalPath := "/etc/ale/config.toml"
	if _, err := os.Open(globalPath); err == nil {
		return globalPath
	}

	return ""
}

// ReadConfigFile reads the config file and merges with DefaultConfig, taking precedence
func ReadConfigFile(cfg *Config, path string) {
	_, err := os.Stat(path)
	if err != nil {
		return
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		logrus.WithError(err).Fatal("unable to read config")
	}
}
