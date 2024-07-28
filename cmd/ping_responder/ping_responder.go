package main

import (
	"flag"
	"strings"

	"github.com/silkeh/matrix_irc_ping_exporter/internal/log"
	"github.com/silkeh/matrix_irc_ping_exporter/irc"
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

	if err := log.Setup(logLevel); err != nil {
		log.Fatal("Invalid loglevel", "level", logLevel, "err", err)
	}

	config.Channels = strings.Split(channelList, ",")
	client, err := irc.NewClient(&config)
	if err != nil {
		log.Fatal("Error connecting to IRC server", "url", config.Server, "err", err)
	}
	client.Loop()
}
