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
			fmt.Fprintf(w, "matrix_irc_ping_seconds{network=\"%s\"} %v\n", d.Room, d.Ping().Seconds())
			fmt.Fprintf(w, "matrix_irc_pong_seconds{network=\"%s\"} %v\n", d.Room, d.Pong().Seconds())
			fmt.Fprintf(w, "matrix_irc_rtt_seconds{network=\"%s\"} %v\n", d.Room, d.RTT().Seconds())
			fmt.Fprintf(w, "matrix_irc_pong_irc_seconds{network=\"%s\"} %v\n", d.Room, d.IRCPong().Seconds())
			fmt.Fprintf(w, "matrix_irc_pong_matrix_seconds{network=\"%s\"} %v\n", d.Room, d.MatrixPong().Seconds())
		}
		fmt.Fprintf(w, "matrix_irc_ping_success{network=\"%s\"} %v\n", n, success)
	}
}

// sendPings sends pings to all configured rooms and returns a map with ping IDs
func (e *Exporter) sendPings() (ids map[string]time.Time) {
	log.Infof("Sending %v pings", len(e.Rooms))

	ids = make(map[string]time.Time, len(e.Rooms))
	for _, roomID := range e.Rooms {
		// Create random ID and register it
		id := util.RandString(idSize)
		ts := time.Now()
		ids[id] = ts

		// Try to send a ping message
		_, err := e.SendPing(roomID, id, ts)
		if err != nil {
			log.Warnf("Error sending ping to room %s: %s", roomID, err)
		}
	}
	return
}

// getDelays returns the sent delays.
func (e *Exporter) getDelays(ids map[string]time.Time) (delays map[string]*matrix.Delay) {
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
			ts, ok := ids[delay.ID]
			if !ok {
				continue
			}

			// Save received delay
			delay.PingTime = ts
			delays[delay.Room] = &delay
			log.Debugf("Received response in %s with RTT of %s", delay.Room, delay.RTT())

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
