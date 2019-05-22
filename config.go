package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// MatrixConfig is used for the configuration of the Matrix client
type MatrixConfig struct {
	Homeserver  string
	User        string
	Token       string
	MessageType string
	Rooms       map[string]string
}

// IRCConfig is used for the configuration of the IRC clients
type IRCConfig struct {
	Server   string
	Nick     string
	Name     string
	SSL      bool
	Channels []string
}

// Config is used for the main configuration
type Config struct {
	IRC    map[string]*IRCConfig
	Matrix *MatrixConfig
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
