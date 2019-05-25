package main

import (
	"flag"
	"strings"

	"github.com/silkeh/matrix_irc_ping_exporter/irc"
	log "github.com/sirupsen/logrus"
)

func main() {
	var config irc.Config
	var channelList, logLevel string

	flag.StringVar(&config.Server, "server", "localhost:6667", "IRC server to connect to")
	flag.StringVar(&config.Nick, "nick", "PingBot", "Nickname to use")
	flag.StringVar(&config.Name, "name", "PingBot", "Real name to use")
	flag.StringVar(&channelList, "channels", "", "Comma separated list of channels to join")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level")
	flag.BoolVar(&config.SSL, "ssl", false, "Use SSL for this connection")
	flag.Parse()

	// Set log level
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid loglevel %q: %s", logLevel, lvl)
	}
	log.SetLevel(lvl)

	config.Channels = strings.Split(channelList, ",")
	client, err := irc.NewClient(&config)
	if err != nil {
		log.Fatalf("Error connecting to %s: %s", config.Server, err)
	}
	client.Loop()
}
