package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/silkeh/matrix_irc_ping_exporter/irc"
	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
)

var (
	ircClients = make(map[string]*irc.Client)
)

func main() {
	var addr, configFile string
	var pingTimeout time.Duration

	flag.StringVar(&addr, "addr", ":9200", "Listen address")
	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file")
	flag.DurationVar(&pingTimeout, "timeout", 30*time.Second, "Ping timeout")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file %s: %s", configFile, err)
	}

	// Create and start a Matrix client
	client, err := matrix.NewClient(config.Matrix)
	if err != nil {
		log.Fatalf("Error connecting to Matrix homeserver: %s", err)
	}
	go client.Sync()

	// Create IRC clients
	for n, conf := range config.IRC {
		ircClients[n], err = irc.NewClient(conf)
		if err != nil {
			log.Fatalf("Error connecting to IRC server %s: %s", conf.Server, err)
		}
		go ircClients[n].Loop()
	}

	// Create exporter
	exporter := NewExporter(client, config.Matrix.Rooms, pingTimeout)

	// Create HTTP server
	http.HandleFunc("/metrics", exporter.MetricsHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
