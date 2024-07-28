package matrix

import (
	"context"
	"fmt"
	"time"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/silkeh/matrix_irc_ping_exporter/ping"

	log "github.com/sirupsen/logrus"
	matrix "maunium.net/go/mautrix"
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
	Rooms        map[id.RoomID]string
	Pings, Pongs chan *ping.Message
	messageType  event.MessageType
}

// Config is used for the configuration of the Matrix client
type Config struct {
	Homeserver  string
	User        string
	Token       string
	MessageType event.MessageType
	Rooms       map[string]id.RoomID
}

// Message represents a Matrix Message
type Message struct {
	MsgType event.MessageType `json:"msgtype"`
	Body    string            `json:"body"`
}

// NewClient returns a configured Matrix Client
func NewClient(config *Config) (c *Client, err error) {
	c = &Client{
		messageType: config.MessageType,
		Rooms:       make(map[id.RoomID]string, len(config.Rooms)),
		Pings:       make(chan *ping.Message, 25),
		Pongs:       make(chan *ping.Message, 25),
	}

	// Add Rooms to map with id/name swapped
	for name, roomID := range config.Rooms {
		c.Rooms[roomID] = name
	}

	// Create the actual Matrix client
	c.Client, err = matrix.NewClient(config.Homeserver, id.UserID(config.User), config.Token)
	if err != nil {
		return
	}

	// Copy a pointer to the syncer for easy access
	c.Syncer = c.Client.Syncer.(*matrix.DefaultSyncer)

	// Register sync/message handler
	c.Syncer.OnEventType(event.NewEventType("m.room.message"), c.messageHandler)

	// Join Rooms
	if len(config.Rooms) > 0 {
		err = c.JoinRooms(context.Background(), config.Rooms)
		if err != nil {
			return
		}
	}

	return
}

// SendPing sends a ping message
func (c *Client) SendPing(ctx context.Context, roomID id.RoomID, pingID string, ts time.Time) (*matrix.RespSendEvent, error) {
	log.Debugf("Sending ping with ID %q to %q", pingID, roomID)

	return c.SendText(ctx, roomID, fmt.Sprintf("%s %s %d", PingMessage, pingID, ts.UnixNano()))
}

// SendText sends a plain text message
func (c *Client) SendText(ctx context.Context, roomID id.RoomID, text string) (*matrix.RespSendEvent, error) {
	return c.SendMessageEvent(ctx, roomID, event.EventMessage,
		Message{
			MsgType: c.messageType,
			Body:    text,
		},
	)
}

// JoinRooms joins a map names to room IDs/aliases
func (c *Client) JoinRooms(ctx context.Context, roomList map[string]id.RoomID) error {
	for _, r := range roomList {
		_, err := c.JoinRoom(ctx, string(r), "", nil)
		if err != nil {
			return err
		}
	}
	return nil
}
