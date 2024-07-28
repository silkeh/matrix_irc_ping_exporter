package matrix

import (
	"context"
	"strconv"
	"strings"
	"time"

	"maunium.net/go/mautrix/event"

	"github.com/silkeh/matrix_irc_ping_exporter/ping"

	log "github.com/sirupsen/logrus"
)

func parseMessage(e *event.Event) (*event.MessageEventContent, bool) {
	err := e.Content.ParseRaw(event.EventMessage)
	if err != nil {
		return &event.MessageEventContent{}, false
	}

	return e.Content.AsMessage(), true
}

// messageHandler handles incoming messages
func (c *Client) messageHandler(ctx context.Context, e *event.Event) {
	now := time.Now()

	// Ignore message if it has no body
	msg, ok := parseMessage(e)
	if !ok || msg.Body == "" {
		log.Debugf("Ignoring message %q without body from %q", e.ID, e.RoomID)
		return
	}

	// Get command from body
	cmd := strings.SplitN(msg.Body, " ", 2)[0]

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
		if msg.MsgType == event.MsgNotice {
			log.Debugf("Ignoring notice message %q from %q", e.ID, e.RoomID)
			return
		}
		err = c.pingHandler(ctx, e, now)
	}

	if err != nil {
		log.Errorf("Error sending %q response: %s", cmd, err)
	}
}

func (c *Client) parseMessage(e *event.Event, received time.Time) *ping.Message {
	// Ignore message if not received in the configured Rooms
	room, ok := c.Rooms[e.RoomID]
	if !ok {
		log.Debugf("Ignoring message %q from unknown room %q", e.ID, e.RoomID)
		return nil
	}

	// Ignore message if not all components are available
	msg, _ := parseMessage(e)
	parts := strings.Split(msg.Body, " ")
	if len(parts) < 3 {
		return nil
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		log.Infof("Received pong with invalid time: %s", msg.Body)
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
