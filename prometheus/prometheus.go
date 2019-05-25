package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
	"github.com/silkeh/matrix_irc_ping_exporter/util"
	log "github.com/sirupsen/logrus"
)

const idSize = 8

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
	// Send ping to all rooms
	ids := e.sendPings()

	// Read all delays
	delays := e.getDelays(ids)

	// Write metrics
	// TODO: use proper exporter functionality for this
	for n, d := range delays {
		success := 0
		if d != nil {
			success = 1
			fmt.Fprintf(w, "matrix_irc_ping_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Ping)/1e9)
			fmt.Fprintf(w, "matrix_irc_pong_delay_seconds{network=\"%s\"} %v\n", d.Room, float64(d.Pong)/1e9)
		}
		fmt.Fprintf(w, "matrix_irc_ping_success{network=\"%s\"} %v\n", n, success)
	}
}

// sendPings sends pings to all configured rooms and returns a map with ping IDs
func (e *Exporter) sendPings() (ids map[string]struct{}) {
	log.Infof("Sending %v pings", len(e.Rooms))

	ids = make(map[string]struct{}, len(e.Rooms))
	for _, id := range e.Rooms {
		// Create random ID and register it
		pid := util.RandString(idSize)
		ids[pid] = struct{}{}

		// Try to send a ping message
		_, err := e.SendPing(id, pid)
		if err != nil {
			log.Warnf("Error sending ping to room %s: %s", id, err)
		}
	}
	return
}

// getDelays returns the sent delays.
func (e *Exporter) getDelays(ids map[string]struct{}) (delays map[string]*matrix.Delay) {
	log.Debugf("Waiting for replies, timeout in %v", e.Timeout)

	// Initialise delays map with nil pointers
	delays = make(map[string]*matrix.Delay, len(e.Rooms))
	for n := range e.Rooms {
		delays[n] = nil
	}

	// Run the rest with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	// Check for incoming messages and return when done,
	// or when the timeout is reached.
	for {
		select {
		case delay := <-e.Delays:
			// Check if this is a ping we sent
			if _, ok := ids[delay.ID]; !ok {
				continue
			}

			// Save received delay
			delays[delay.Room] = &delay

			// Stop when everything has been received
			if len(delays) == len(e.Rooms) {
				return
			}
		case <-ctx.Done():
			log.Info("Timed out waiting for replies.")
			return
		}
	}
}
