package matrix

import (
	"html"
	"strings"
	"text/template"
	"time"

	matrix "github.com/matrix-org/gomatrix"
)

const (
	pingCommandResponseBody          = "{{.Sender}}: Pong! (ping {{.Message}} {{.Duration}} to arrive)"
	pingCommandResponseFormattedBody = "<a href='https://matrix.to/#/{{.Sender}}'>{{.Sender}}</a>: Pong! " +
		"(<a href='https://matrix.to/#/{{.RoomID}}/{{.ID}}'>ping</a> {{.Message}} {{.Duration}} to arrive)"
)

var (
	pingTextTemplate = template.Must(template.New("ping").Parse(pingCommandResponseBody))
	pingHTMLTemplate = template.Must(template.New("ping").Parse(pingCommandResponseFormattedBody))
)

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
	_ = pingTextTemplate.Execute(&plain, event)
	_ = pingHTMLTemplate.Execute(&formatted, event)

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
