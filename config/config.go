package config

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/kardianos/osext"
	"github.com/kelseyhightower/envconfig"
)

// Config struct holds the current configuration
type Config struct {
	// Server settings
	Address string `yaml:"address" envconfig:"address"`
	Port    int    `yaml:"port" envconfig:"port"`

	// Logger settings
	LogLevel  string `yaml:"loglevel" envconfig:"loglevel"`
	LogFormat string `yaml:"logformat" envconfig:"logformat"`

	// GCS Bucket
	Database DBConfig `yaml:"database"`

	// Service settings
	// - Owner of the service. For example the team running it.
	//   Defaulted to the current user.
	Owner string `yaml:"owner" envconfig:"owner"`
}

// DBConfig represents the Database Configuration object
type DBConfig struct {
	Type    string `yaml:"type"`
	Project string `yaml:"project"`
}

// Initialize a new Config
func Initialize(configFile string) *Config {
	cfg := DefaultConfig()
	ReadConfigFile(cfg, getConfigFilePath(configFile))
	ReadEnvironment(cfg)

	return cfg
}

// DefaultConfig returns a Config struct with default values
func DefaultConfig() *Config {
	return &Config{
		Address: "0.0.0.0",
		Port:    7654,

		LogLevel:  "debug",
		LogFormat: "text",

		Database: DBConfig{
			Type: "file",
		},

		Owner: os.Getenv("USER"),
	}
}

// getConfigFilePath returns the location of the config file in order of priority:
// 1 ) --config commandline flag
// 1 ) File in same directory as the executable
// 2 ) Global file in /etc/ale/config.yml
func getConfigFilePath(configPath string) string {
	if configPath != "" {
		path := fmt.Sprintf("%s/config.yml", configPath)
		if _, err := os.Open(path); err == nil {
			return path
		}
		panic(fmt.Sprintf("Unable to open %s.", path))
	}
	path, _ := osext.ExecutableFolder()
	path = fmt.Sprintf("%s/config.yml", path)
	if _, err := os.Open(path); err == nil {
		return path
	}
	globalPath := "/etc/ale/config.yml"
	if _, err := os.Open(globalPath); err == nil {
		return globalPath
	}

	return ""
}

// ReadConfigFile reads the config file and merges with DefaultConfig, taking precedence
func ReadConfigFile(cfg *Config, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	configFile, _ := ioutil.ReadAll(file)
	yaml.Unmarshal(configFile, cfg)
}

// ReadEnvironment takes precedence over any configs set with settings provided
// from the environment
func ReadEnvironment(cfg *Config) {
	envconfig.Process("ALE", cfg)
}
