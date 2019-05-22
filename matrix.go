package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/matrix-org/gomatrix"
	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
)

// PingResponse contains the expected response to a ping message
const PingResponse = "pong"

// Delay represents a delay for a room.
type Delay struct {
	Room       string
	Ping, Pong time.Duration
}

// MatrixClient represents a Matrix client that collects ping responses.
type MatrixClient struct {
	*matrix.Client
	rooms     map[string]string
	responses chan Delay
}

// NewMatrixClient starts a new matrix client
func NewMatrixClient(config *MatrixConfig) (client *MatrixClient, err error) {
	client = &MatrixClient{
		rooms:     make(map[string]string, len(config.Rooms)),
		responses: make(chan Delay, 25),
	}

	// Add rooms to map with key/value swapped
	for k, v := range config.Rooms {
		client.rooms[v] = k
	}

	// Create a new client
	client.Client, err = matrix.NewClient(config.Homeserver, config.User, config.Token, config.MessageType)
	if err != nil {
		return
	}

	// Register sync/message handler
	client.Syncer.OnEventType("m.room.message", client.messageHandler)

	// Join rooms
	if len(config.Rooms) > 0 {
		err = client.joinRooms(config.Rooms)
		if err != nil {
			return
		}
	}

	// Start syncing
	go client.sync()

	return
}

// joinRooms joins a list of room IDs or aliases
func (client *MatrixClient) joinRooms(roomList map[string]string) error {
	for _, r := range roomList {
		_, err := client.JoinRoom(r, "", nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// messageHandler handles ping responses
func (client *MatrixClient) messageHandler(e *gomatrix.Event) {
	// Ignore message if not received in the configured rooms
	room, ok := client.rooms[e.RoomID]
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
	if len(parts) < 3 {
		return
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(parts[1], 0, 64)
	if err != nil {
		log.Printf("Received pong with invalid time: %s", text)
		return
	}

	// Parse initial delay
	lag, err := strconv.ParseInt(parts[2], 0, 64)
	if err != nil {
		log.Printf("Received pong with invalid delay: %s", text)
		return
	}

	// Queue response
	client.responses <- Delay{
		Room: room,
		Ping: time.Duration(lag),
		Pong: time.Since(time.Unix(0, ts)),
	}

	return
}

// sync runs a never ending Matrix sync
func (client *MatrixClient) sync() {
	for {
		err := client.Client.Sync()
		if err != nil {
			log.Printf("Sync error: %s", err)
		}
		//time.Sleep(0 * time.Second)
	}
}
