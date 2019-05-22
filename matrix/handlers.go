package matrix

import (
	"log"
	"strconv"
	"strings"
	"time"

	matrix "github.com/matrix-org/gomatrix"
)

// MessageHandler handles incoming ping responses
func (c *Client) messageHandler(e *matrix.Event) {
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
		log.Printf("Received pong with invalid time: %s", text)
		return
	}

	// Parse initial delay
	lag, err := strconv.ParseInt(parts[3], 0, 64)
	if err != nil {
		log.Printf("Received pong with invalid delay: %s", text)
		return
	}

	// Queue response
	c.Delays <- Delay{
		Room: room,
		ID: parts[1],
		Ping: time.Duration(lag),
		Pong: time.Since(time.Unix(0, ts)),
	}

	return
}

// Sync runs a never ending Matrix sync
func (c *Client) Sync() {
	for {
		err := c.Client.Sync()
		if err != nil {
			log.Printf("Sync error: %s", err)
		}
	}
}
