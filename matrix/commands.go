package matrix

import (
	"context"
	"fmt"
	"html"
	"strings"
	"text/template"
	"time"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
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
	*event.Event
	Message  string
	Duration time.Duration
}

type pongData struct {
	Milliseconds int64      `json:"ms"`
	From         string     `json:"from"`
	Ping         id.EventID `json:"ping"`
}

type pingMessage struct {
	event.MessageEventContent
	Pong pongData `json:"pong"`
}

func (c *Client) pingHandler(ctx context.Context, e *event.Event, received time.Time) error {
	ev := &pingEvent{Event: e}

	// Set message from body
	args := getArgs(e, 1)
	if len(args) == 0 {
		ev.Message = "took"
	} else {
		body := args[0]
		if len(body) > 32 {
			body = body[:32]
		}
		ev.Message = `"` + body + `" took`
	}

	// Calculate time difference
	ev.Duration = received.Sub(time.Unix(0, ev.Timestamp*1e6))

	// Create response
	var plain, formatted strings.Builder
	_ = pingTemplate.ExecuteTemplate(&plain, "text", ev)
	_ = pingTemplate.ExecuteTemplate(&formatted, "html", ev)

	// Create response
	response := &pingMessage{
		MessageEventContent: event.MessageEventContent{
			MsgType:       c.messageType,
			Body:          html.EscapeString(plain.String()),
			Format:        HTMLFormat,
			FormattedBody: formatted.String(),
		},
		Pong: pongData{
			Milliseconds: ev.Duration.Milliseconds(),
			From:         ev.Sender.Homeserver(),
			Ping:         ev.ID,
		},
	}

	_, err := c.SendMessageEvent(ctx, e.RoomID, event.EventMessage, response)
	return err
}

func getArgs(e *event.Event, n int) []string {
	msg, _ := parseMessage(e)
	return strings.SplitN(msg.Body, " ", n+1)[1:]
}

func formatDuration(d time.Duration) string {
	switch {
	case d < 10*time.Second:
		return fmt.Sprintf("%d ms", d.Milliseconds())
	case d < 1*time.Minute:
		return fmt.Sprintf("%.1f second", d.Seconds())
	default:
		return fmt.Sprintf("%6s", d.Truncate(time.Second))
	}
}
