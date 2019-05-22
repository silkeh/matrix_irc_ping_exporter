package matrix

import (
	"time"

	matrix "github.com/matrix-org/gomatrix"
)

const (
	// PingMessage contains the prefix for a ping message
	PingMessage = "ping"

	// PingResponse contains the expected response prefix to a ping message
	PingResponse = "pong"
)

// Client represents a Matrix Client that can be used as a ping sender.
type Client struct {
	*matrix.Client
	Syncer      *matrix.DefaultSyncer
	Rooms       map[string]string
	Delays      chan Delay
	messageType string
}

// Config is used for the configuration of the Matrix client
type Config struct {
	Homeserver  string
	User        string
	Token       string
	MessageType string
	Rooms       map[string]string
}

// Delay represents a delay for a room.
type Delay struct {
	Room       string
	Ping, Pong time.Duration
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
		Delays:      make(chan Delay, 25),
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
