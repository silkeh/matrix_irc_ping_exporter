package matrix

import (
	"github.com/silkeh/matrix_irc_ping_exporter/ping"
	"strconv"
	"strings"
	"time"

	matrix "github.com/matrix-org/gomatrix"
	log "github.com/sirupsen/logrus"
)

// MessageHandler handles incoming ping responses
func (c *Client) messageHandler(e *matrix.Event) {
	now := time.Now()

	// Ignore message if not received in the configured Rooms
	room, ok := c.Rooms[e.RoomID]
	if !ok {
		return
	}

	// Parse message
	msg := c.parseMessage(e)
	if msg == nil {
		return
	}

	// Complete message
	msg.Room = room
	msg.Received = now
	log.Debugf("Received %s message with ID %q from %q", msg.Kind, msg.ID, e.RoomID)

	switch {
	case msg.Kind == PingMessage:
		c.Pings <- msg
	case msg.Kind == PingResponse:
		c.Pongs <- msg
	}
}

func (c *Client) parseMessage(e *matrix.Event) *ping.Message {
	// Ignore message if it has no body
	text, ok := e.Body()
	if !ok {
		return nil
	}

	// Ignore message if not all components are available
	parts := strings.Split(text, " ")
	if len(parts) < 3 {
		return nil
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		log.Infof("Received pong with invalid time: %s", text)
		return nil
	}

	// Assemble message
	return &ping.Message{
		Kind:   parts[0],
		ID:     parts[1],
		Sent:   time.Unix(0, ts),
		Matrix: time.Unix(0, e.Timestamp*1e6),
	}
}

// Sync runs a never ending Matrix sync
func (c *Client) Sync() {
	for {
		err := c.Client.Sync()
		if err != nil {
			log.Errorf("Sync error: %s", err)
		}
	}
}
