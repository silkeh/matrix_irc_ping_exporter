package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/silkeh/matrix_irc_ping_exporter/irc"
)

func setBoolFromEnv(target *bool, env string) {
	if str := os.Getenv(env); str != "" {
		*target, _ = strconv.ParseBool(str)
	}
}

func setStringFromEnv(target *string, env string) {
	if str := os.Getenv(env); str != "" {
		*target = str
	}
}

var (
	ircClients = make(map[string]*irc.Client)
)

func main() {
	var addr, configFile string

	flag.StringVar(&addr, "addr", ":9200", "Listen address")
	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file %s: %s", configFile, err)
	}

	// Create and start aatrix client
	client, err := NewMatrixClient(config.Matrix)
	if err != nil {
		log.Fatalf("Error connecting to Matrix homeserver: %s", err)
	}

	// Create IRC clients
	for n, conf := range config.IRC {
		ircClients[n], err = irc.NewClient(conf.Server, conf.Nick, conf.Name, conf.SSL, conf.Channels)
		if err != nil {
			log.Fatalf("Error connecting to IRC server %s: %s", conf.Server, err)
		}
		go ircClients[n].Loop()
	}

	// Create HTTP server
	http.HandleFunc("/metrics", client.metricsHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
