package matrix

import (
	"fmt"
	"github.com/silkeh/matrix_irc_ping_exporter/ping"
	"time"

	matrix "github.com/matrix-org/gomatrix"
	log "github.com/sirupsen/logrus"
)

const (
	// PingMessage contains the prefix for a ping message
	PingMessage = "ping"

	// PingResponse contains the expected response prefix to a ping message
	PingResponse = "pong"

	// PingCommand is the prefix for a ping command
	PingCommand = "!ping"

	// HTMLFormat is the HTML format used for HTML formatted messages
	HTMLFormat = "org.matrix.custom.html"
)

// Client represents a Matrix Client that can be used as a ping sender.
type Client struct {
	*matrix.Client
	Syncer       *matrix.DefaultSyncer
	Rooms        map[string]string
	Pings, Pongs chan *ping.Message
	messageType  string
}

// Config is used for the configuration of the Matrix client
type Config struct {
	Homeserver  string
	User        string
	Token       string
	MessageType string
	Rooms       map[string]string
}

// Message represents a Matrix Message
type Message struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// NewClient returns a configured Matrix Client
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{
		messageType: config.MessageType,
		Rooms:       make(map[string]string, len(config.Rooms)),
		Pings:       make(chan *ping.Message, 25),
		Pongs:       make(chan *ping.Message, 25),
	}

	// Add Rooms to map with id/name swapped
	for name, id := range config.Rooms {
		c.Rooms[id] = name
	}

	// Create the actual Matrix client
	c.Client, err = matrix.NewClient(config.Homeserver, config.User, config.Token)
	if err != nil {
		return
	}

	// Copy a pointer to the syncer for easy access
	c.Syncer = c.Client.Syncer.(*matrix.DefaultSyncer)

	// Register sync/message handler
	c.Syncer.OnEventType("m.room.message", c.messageHandler)

	// Join Rooms
	if len(config.Rooms) > 0 {
		err = c.JoinRooms(config.Rooms)
		if err != nil {
			return
		}
	}

	return
}

// SendPing sends a ping message
func (c *Client) SendPing(roomID, pingID string, ts time.Time) (*matrix.RespSendEvent, error) {
	log.Debugf("Sending ping with ID %q to %q", pingID, roomID)

	return c.SendText(roomID, fmt.Sprintf("%s %s %d", PingMessage, pingID, ts.UnixNano()))
}

// SendText sends a plain text message
func (c *Client) SendText(roomID, text string) (*matrix.RespSendEvent, error) {
	return c.SendMessageEvent(roomID, "m.room.message",
		Message{
			MsgType: c.messageType,
			Body:    text,
		},
	)
}

// JoinRooms joins a map names to room IDs/aliases
func (c *Client) JoinRooms(roomList map[string]string) error {
	for _, r := range roomList {
		_, err := c.JoinRoom(r, "", nil)
		if err != nil {
			return err
		}
	}
	return nil
}
