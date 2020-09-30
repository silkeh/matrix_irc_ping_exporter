package matrix

import (
	"fmt"
	"html"
	"strings"
	"text/template"
	"time"

	matrix "github.com/matrix-org/gomatrix"
)

const (
	pingCommandResponseBody          = "{{.Sender}}: Pong! (ping {{.Message}} {{formatDuration .Duration}} to arrive)"
	pingCommandResponseFormattedBody = "<a href='https://matrix.to/#/{{.Sender}}'>{{.Sender}}</a>: Pong! " +
		"(<a href='https://matrix.to/#/{{.RoomID}}/{{.ID}}'>ping</a> {{.Message}} {{formatDuration .Duration}} to arrive)"
)

var pingTemplate *template.Template

func init() {
	pingTemplate = template.New("").Funcs(template.FuncMap{"formatDuration": formatDuration})
	pingTemplate = template.Must(pingTemplate.New("text").Parse(pingCommandResponseBody))
	pingTemplate = template.Must(pingTemplate.New("html").Parse(pingCommandResponseFormattedBody))
}

type pingEvent struct {
	*matrix.Event
	Message  string
	Duration time.Duration
}

type pongData struct {
	Milliseconds int64  `json:"ms"`
	From         string `json:"from"`
	Ping         string `json:"ping"`
}

type pingMessage struct {
	matrix.HTMLMessage
	Pong pongData `json:"pong"`
}

func (c *Client) pingHandler(e *matrix.Event, received time.Time) error {
	event := &pingEvent{Event: e}

	// Set message from body
	args := getArgs(e, 1)
	if len(args) == 0 {
		event.Message = "took"
	} else {
		body := args[0]
		if len(body) > 32 {
			body = body[:32]
		}
		event.Message = `"` + body + `" took`
	}

	// Calculate time difference
	event.Duration = received.Sub(time.Unix(0, event.Timestamp*1e6))

	// Create response
	var plain, formatted strings.Builder
	_ = pingTemplate.ExecuteTemplate(&plain, "text", event)
	_ = pingTemplate.ExecuteTemplate(&formatted, "html", event)

	// Create response
	response := &pingMessage{
		HTMLMessage: matrix.HTMLMessage{
			MsgType:       c.messageType,
			Body:          html.EscapeString(plain.String()),
			Format:        HTMLFormat,
			FormattedBody: formatted.String(),
		},
		Pong: pongData{
			Milliseconds: event.Duration.Milliseconds(),
			From: strings.SplitN(event.Sender, ":", 2)[1],
			Ping: event.ID,
		},
	}

	_, err := c.SendMessageEvent(e.RoomID, "m.room.message", response)
	return err
}

func getArgs(e *matrix.Event, n int) []string {
	body, _ := e.Body()
	return strings.SplitN(body, " ", n+1)[1:]
}

func formatDuration(d time.Duration) string {
	switch {
	case d < 10 * time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	case d < 1 * time.Minute:
		return fmt.Sprintf("%.1f second", d.Seconds())
	default:
		return fmt.Sprintf("%6s", d.Truncate(time.Second))
	}
}
