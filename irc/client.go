package irc

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	irc "github.com/thoj/go-ircevent"
	irclib "gopkg.in/sorcix/irc.v2"
)

// Client is a simple IRC pong client
type Client struct {
	*irc.Connection
	Channels []string
}

// Config is the configuration for a Client.
type Config struct {
	Server   string
	Nick     string
	Name     string
	SSL      bool
	Channels []string
}

// NewClient creates and connects a simple IRC pong client
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{
		Channels:   config.Channels,
		Connection: irc.IRC(config.Nick, config.Name),
	}

	// Catch invalid config
	if c.Connection == nil {
		return nil, fmt.Errorf("invalid IRC name or realname: %q, %q", config.Nick, config.Name)
	}

	// Configure the client
	c.UseTLS = config.SSL

	// Register callbacks
	c.AddCallback(irclib.RPL_WELCOME, c.onConnect)
	c.AddCallback(irclib.PRIVMSG, c.onPrivMsg)

	// Connect
	err = c.Connect(config.Server)
	if err != nil {
		return
	}

	return
}

// onConnect handles what should happen after a connection has been established
func (c *Client) onConnect(e *irc.Event) {
	for _, ch := range c.Channels {
		c.Join(ch)
	}
}

// onPrivMsg handles incoming messages
func (c *Client) onPrivMsg(e *irc.Event) {
	if len(e.Arguments) != 2 {
		return
	}

	channel := e.Arguments[0]
	if channel == c.GetNick() {
		channel = e.Nick
	}

	msg := strings.TrimSpace(e.Message())
	if strings.HasPrefix(msg, "ping") {
		c.Privmsg(channel, createResponse(msg))
	}
}

// createResponse creates the correct response for an incoming ping message.
func createResponse(msg string) string {
	now := time.Now()
	str := fmt.Sprintf("pong %v", now.UnixNano())

	// Check if a timestamp was given, and calculate the difference.
	// This difference is appended to the return message.
	parts := strings.Split(msg, " ")
	if len(parts) >= 2 {
		ts, err := strconv.ParseInt(parts[1], 0, 64)
		if err == nil {
			diff := now.Sub(time.Unix(0, ts))
			str += fmt.Sprintf(" %d %s", diff, diff)
		}
	}

	return str
}
