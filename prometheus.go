package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

const timeout = 60 * time.Second

func (client *MatrixClient) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: clear receiver channel

	// Send ping to all rooms
	for id := range client.rooms {
		_, err := client.SendText(id, fmt.Sprintf("ping %d", time.Now().UnixNano()))
		if err != nil {
			log.Printf("Error sending ping to room %s: %s", id, err)
		}
	}

	// Read all responses
	delays := client.getDelays()

	// Write metrics
	// TODO: use proper exporter functionality for this
	for _, d := range delays {
		fmt.Fprintf(w, "matrix_irc_ping_timeout{network=\"%s\"} %v\n", d.Room, 0)
		fmt.Fprintf(w, "matrix_irc_ping_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Ping)/1e9)
		fmt.Fprintf(w, "matrix_irc_pong_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Pong)/1e9)
	}
}

func (client *MatrixClient) getDelays() (delays map[string]Delay) {
	// Initialise delays with timeout values
	delays = make(map[string]Delay, len(client.rooms))
	for _, n := range client.rooms {
		delays[n] = Delay{Room: n, Ping: timeout, Pong: timeout}
	}

	// Run the rest with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Check for incoming messages and return when done,
	// or when the timeout is reached.
	for {
		select {
		case delay := <-client.responses:
			delays[delay.Room] = delay
			if len(delays) == len(client.rooms) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
