package main

import (
	"flag"
	"github.com/silkeh/matrix_irc_ping_exporter/irc"
	"log"
	"strings"
)

func main() {
	var config irc.Config
	var channelList string

	flag.StringVar(&config.Server, "server", "localhost:6667", "IRC server to connect to")
	flag.StringVar(&config.Nick, "nick", "PingBot", "Nickname to use")
	flag.StringVar(&config.Name, "name", "PingBot", "Real name to use")
	flag.StringVar(&channelList, "channels", "", "Comma separated list of channels to join")
	flag.BoolVar(&config.SSL, "ssl", false, "Use SSL for this connection")
	flag.Parse()

	config.Channels = strings.Split(channelList, ",")
	client, err := irc.NewClient(&config)
	if err != nil {
		log.Fatalf("Error connecting to %s: %s", config.Server, err)
	}
	client.Loop()
}
