package prometheus

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"maunium.net/go/mautrix/id"

	"github.com/silkeh/matrix_irc_ping_exporter/matrix"
	"github.com/silkeh/matrix_irc_ping_exporter/ping"
	"github.com/silkeh/matrix_irc_ping_exporter/util"
)

const idSize = 8

// Exporter is a Prometheus exporter for Matrix-IRC ping metrics.
type Exporter struct {
	*matrix.Client
	Rooms   map[string]id.RoomID
	Timeout time.Duration
}

// NewExporter returns a configured ping metrics exporter.
func NewExporter(client *matrix.Client, rooms map[string]id.RoomID, timeout time.Duration) *Exporter {
	return &Exporter{
		Client:  client,
		Rooms:   rooms,
		Timeout: timeout,
	}
}

// MetricsHandler is an HTTP handler that collects metrics.
func (e *Exporter) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), e.Timeout)
	defer cancel()

	// Send ping to all rooms
	ids := e.sendPings(ctx)

	// Read all delays
	delays := e.getDelays(ctx, ids)

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
func (e *Exporter) sendPings(ctx context.Context) (ids map[string]time.Time) {
	slog.Info("Sending pings", "count", len(e.Rooms))

	ids = make(map[string]time.Time, len(e.Rooms))
	for _, roomID := range e.Rooms {
		// Create random ID and register it
		id := util.RandString(idSize)
		ts := time.Now()
		ids[id] = ts

		// Try to send a ping message
		_, err := e.SendPing(ctx, roomID, id, ts)
		if err != nil {
			slog.Warn("Error sending ping", "room_id", roomID, "err", err)
		}
	}
	return
}

// getDelays returns the sent delays.
func (e *Exporter) getDelays(ctx context.Context, ids map[string]time.Time) (delays map[string]*ping.Delay) {
	slog.Debug("Waiting for replies", "timeout", e.Timeout)

	// Initialise delays map with nil pointers
	delays = make(map[string]*ping.Delay, len(e.Rooms))
	for n := range e.Rooms {
		delays[n] = new(ping.Delay)
	}

	// Check for incoming messages and return when done,
	// or when the timeout is reached.
	pingCount := 0
	pongCount := 0
	for {
		select {
		case msg := <-e.Pings:
			// Check if this is a ping we sent
			ts, ok := ids[msg.ID]
			if !ok {
				slog.Debug("Ignoring ping", "ping_id", msg.ID)
				continue
			}

			// Store message as ping
			msg.Sent = ts
			delays[msg.Room].Ping = msg

			slog.Debug("Received ping", "room_id", msg.Room, "delay", msg.ToMatrix())

			// Stop when everything has been received
			pingCount++
			if pingCount == len(e.Rooms) && pongCount == len(e.Rooms) {
				return
			}

		case msg := <-e.Pongs:
			// Check if this is a ping we sent
			if _, ok := ids[msg.ID]; !ok {
				slog.Debug("Ignoring pong", "ping_id", msg.ID)
				continue
			}

			// Store message as pong
			delays[msg.Room].Pong = msg

			slog.Debug("Received pong", "room_id", msg.Room, "total_delay", msg.Total())

			// Stop when everything has been received
			pongCount++
			if pingCount == len(e.Rooms) && pongCount == len(e.Rooms) {
				return
			}

		case <-ctx.Done():
			slog.Info("Timed out waiting for replies.")
			return
		}
	}
}
