package prometheus

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
)

const timeout = 30 * time.Second

// Exporter is a Prometheus exporter for Matrix-IRC ping metrics.
type Exporter struct {
	*matrix.Client
	Rooms   map[string]string
	Timeout time.Duration
}

// NewExporter returns a configured ping metrics exporter..
func NewExporter(client *matrix.Client, rooms map[string]string, timeout time.Duration) *Exporter {
	return &Exporter{
		Client:  client,
		Rooms:   rooms,
		Timeout: timeout,
	}
}

// MetricsHandler is an HTTP handler that collects metrics.
func (e *Exporter) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: clear receiver channel

	// Send ping to all rooms
	for _, id := range e.Rooms {
		_, err := e.SendPing(id)
		if err != nil {
			log.Printf("Error sending ping to room %s: %s", id, err)
		}
	}

	// Read all delays
	delays := e.getDelays()

	// Write metrics
	// TODO: use proper exporter functionality for this
	for _, d := range delays {
		pingTimeout := 0
		if d.Ping == timeout || d.Pong == timeout {
			pingTimeout = 1
		}
		fmt.Fprintf(w, "matrix_irc_ping_timeout{network=\"%s\"} %v\n", d.Room, pingTimeout)
		fmt.Fprintf(w, "matrix_irc_ping_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Ping)/1e9)
		fmt.Fprintf(w, "matrix_irc_pong_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Pong)/1e9)
	}
}

// getDelays returns the sent delays.
func (e *Exporter) getDelays() (delays map[string]matrix.Delay) {
	// Initialise delays with timeout values
	delays = make(map[string]matrix.Delay, len(e.Rooms))
	for n := range e.Rooms {
		delays[n] = matrix.Delay{Room: n, Ping: timeout, Pong: timeout}
	}

	// Run the rest with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	// Check for incoming messages and return when done,
	// or when the timeout is reached.
	for {
		select {
		case delay := <-e.Delays:
			delays[delay.Room] = delay
			if len(delays) == len(e.Rooms) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
