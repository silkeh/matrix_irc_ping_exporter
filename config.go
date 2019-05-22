package main

import (
	"io/ioutil"

	"github.com/silkeh/matrix_irc_ping_exporter/irc"
	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
	"gopkg.in/yaml.v2"
)

// Config is used for the main configuration
type Config struct {
	IRC    map[string]*irc.Config
	Matrix *matrix.Config
}

// loadConfig loads configuration
func loadConfig(path string) (config *Config, err error) {
	config = new(Config)

	// Get configuration data
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	// Parse configuration
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return
	}

	// Defaults
	if config.Matrix.MessageType == "" {
		config.Matrix.MessageType = "m.notice"
	}

	return
}
