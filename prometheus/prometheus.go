package prometheus

import (
	"context"
	"fmt"
	"github.com/silkeh/matrix_irc_ping_exporter/ping"
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

		// Client to matrix
		if d.Ping != nil {
			// Received time is 'our' received time, ergo: loopback time.
			d.Ping.Matrix = d.Ping.Received

			fmt.Fprintf(w, "matrix_irc_ping_matrix_delay_seconds{network=\"%s\"} %v\n", d.Ping.Room, d.Ping.ToMatrix().Seconds())
		}

		// Complete path
		if d.Ping != nil && d.Pong != nil {
			success = 1

			// We now from the ping reply when the ping was actually received
			d.Ping.Received = d.Pong.Sent

			// Matrix to IRC
			fmt.Fprintf(w, "matrix_irc_ping_delay_seconds{network=\"%s\"} %v\n", d.Ping.Room, d.Ping.Total().Seconds())
			fmt.Fprintf(w, "matrix_irc_ping_irc_delay_seconds{network=\"%s\"} %v\n", d.Ping.Room, d.Ping.FromMatrix().Seconds())

			// IRC to Matrix
			fmt.Fprintf(w, "matrix_irc_pong_delay_seconds{network=\"%s\"} %v\n", d.Pong.Room, d.Pong.Total().Seconds())
			fmt.Fprintf(w, "matrix_irc_pong_matrix_delay_seconds{network=\"%s\"} %v\n", d.Pong.Room, d.Pong.FromMatrix().Seconds())
			fmt.Fprintf(w, "matrix_irc_pong_irc_delay_seconds{network=\"%s\"} %v\n", d.Pong.Room, d.Pong.ToMatrix().Seconds())

			// Complete path
			fmt.Fprintf(w, "matrix_irc_rtt_seconds{network=\"%s\"} %v\n", n, d.RTT().Seconds())
		}

		// Success status
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
func (e *Exporter) getDelays(ids map[string]time.Time) (delays map[string]*ping.Delay) {
	log.Debugf("Waiting for replies, timeout in %v", e.Timeout)

	// Initialise delays map with nil pointers
	delays = make(map[string]*ping.Delay, len(e.Rooms))
	for n := range e.Rooms {
		delays[n] = new(ping.Delay)
	}

	// Run the rest with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	// Check for incoming messages and return when done,
	// or when the timeout is reached.
	pongCount := 0
	for {
		select {
		case msg := <-e.Pings:
			// Check if this is a ping we sent
			ts, ok := ids[msg.ID]
			if !ok {
				log.Debugf("Ignoring ping with ID %s", msg.ID)
				continue
			}

			// Store message as ping
			msg.Sent = ts
			delays[msg.Room].Ping = msg
			log.Debugf("Received ping for %s with Matrix delay of %s",
				msg.Room, msg.ToMatrix())

		case msg := <-e.Pongs:
			// Check if this is a ping we sent
			if _, ok := ids[msg.ID]; !ok {
				log.Debugf("Ignoring pong with ID %s", msg.ID)
				continue
			}

			// Store message as pong
			delays[msg.Room].Pong = msg
			log.Debugf("Received response for %s with RTT of %s",
				msg.Room, delays[msg.Room].RTT())

			// Stop when everything has been received
			pongCount++
			if pongCount == len(e.Rooms) {
				return
			}

		case <-ctx.Done():
			log.Info("Timed out waiting for replies.")
			return
		}
	}
}
