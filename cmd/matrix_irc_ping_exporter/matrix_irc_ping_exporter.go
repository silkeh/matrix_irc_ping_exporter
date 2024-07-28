package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/silkeh/matrix_irc_ping_exporter/internal/log"
	"github.com/silkeh/matrix_irc_ping_exporter/irc"
	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
	"github.com/silkeh/matrix_irc_ping_exporter/prometheus"
)

var ircClients = make(map[string]*irc.Client)

func main() {
	var addr, configFile, logLevel string
	var pingTimeout time.Duration

	flag.StringVar(&addr, "addr", ":9200", "Listen address")
	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level")
	flag.DurationVar(&pingTimeout, "timeout", 60*time.Second, "Ping timeout")
	flag.Parse()

	if err := log.Setup(logLevel); err != nil {
		log.Fatal("Invalid loglevel", "level", logLevel, "err", err)
	}

	// Load configuration
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatal("Error loading config file", "path", configFile, "err", err)
	}

	// Create and start a Matrix client
	client, err := matrix.NewClient(config.Matrix)
	if err != nil {
		log.Fatal("Error connecting to Matrix homeserver", "err", err)
	}
	go client.Sync()

	// Create IRC clients
	for n, conf := range config.IRC {
		ircClients[n], err = irc.NewClient(conf)
		if err != nil {
			log.Fatal("Error connecting to IRC server", "url", conf.Server, "err", err)
		}
		go ircClients[n].Loop()
	}

	// Create exporter
	exporter := prometheus.NewExporter(client, config.Matrix.Rooms, pingTimeout)

	// Create HTTP server
	http.HandleFunc("/metrics", exporter.MetricsHandler)
	log.Fatal("Listen error", "err", http.ListenAndServe(addr, nil))
}
