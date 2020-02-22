package matrix

import (
	"github.com/silkeh/matrix_irc_ping_exporter/ping"
	"strconv"
	"strings"
	"time"

	matrix "github.com/matrix-org/gomatrix"
	log "github.com/sirupsen/logrus"
)

// messageHandler handles incoming messages
func (c *Client) messageHandler(e *matrix.Event) {
	now := time.Now()

	// Ignore message if it has no body
	body, ok := e.Body()
	if !ok {
		log.Debugf("Ignoring message %q without body from %q", e.ID, e.RoomID)
		return
	}

	// Get command from body
	cmd := strings.SplitN(body, " ", 2)[0]

	log.Debugf("Received %q message from %q", cmd, e.RoomID)

	var err error
	switch cmd {
	case PingMessage, PingResponse:
		msg := c.parseMessage(e, now)
		switch {
		case msg == nil:
			// ignore
		case cmd == PingMessage:
			c.Pings <- msg
		case cmd == PingResponse:
			c.Pongs <- msg
		}
	case PingCommand:
		// Ignore notice messages
		if t, ok := e.MessageType(); ok && t == "m.notice" {
			log.Debugf("Ignoring notice message %q from %q", e.ID, e.RoomID)
			return
		}
		err = c.pingHandler(e, now)
	}

	if err != nil {
		log.Errorf("Error sending %q response: %s", cmd, err)
	}
}

func (c *Client) parseMessage(e *matrix.Event, received time.Time) *ping.Message {
	// Ignore message if not received in the configured Rooms
	room, ok := c.Rooms[e.RoomID]
	if !ok {
		log.Debugf("Ignoring message %q from unknown room %q", e.ID, e.RoomID)
		return nil
	}

	// Ignore message if not all components are available
	text, _ := e.Body()
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
		Kind:     parts[0],
		ID:       parts[1],
		Sent:     time.Unix(0, ts),
		Matrix:   time.Unix(0, e.Timestamp*1e6),
		Room:     room,
		Received: received,
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
