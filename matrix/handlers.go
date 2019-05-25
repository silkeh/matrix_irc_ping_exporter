package matrix

import (
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

	// Ignore message if the body does not start with the right response
	text, ok := e.Body()
	if !ok || !strings.HasPrefix(text, PingResponse) {
		return
	}

	// Ignore message if not all components are available
	parts := strings.Split(text, " ")
	if len(parts) <= 4  {
		return
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		log.Infof("Received pong with invalid time: %s", text)
		return
	}

	// Queue response
	log.Debugf("Received pong with ID %q from %q", parts[1], e.RoomID)
	c.Delays <- Delay{
		Room: room,
		ID: parts[1],
		PongTime: time.Unix(0, ts),
		MatrixTime: time.Unix(0, e.Timestamp * 1e6),
		ReceiveTime: now,
	}

	return
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
