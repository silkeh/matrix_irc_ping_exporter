package ping

import "time"

// Message represents a ping or pong message.
// It is sent from a client, arrives on matrix, and is received by another client.
type Message struct {
	Kind, Room, ID         string
	Sent, Matrix, Received time.Time
}

// ToMatrix returns the delay from the sender to matrix.
func (m *Message) ToMatrix() time.Duration {
	return m.Matrix.Sub(m.Sent)
}

// FromMatrix returns the delay from matrix to the receiver.
func (m *Message) FromMatrix() time.Duration {
	return m.Received.Sub(m.Matrix)
}

// Total returns the total delay from the sender to the receiver.
func (m *Message) Total() time.Duration {
	return m.Received.Sub(m.Sent)
}

// Delay represents a delay for a room.
type Delay struct {
	Ping, Pong *Message
}

// RTT returns the delay between the sending of the ping, and the reception of the ping reply.
func (d *Delay) RTT() time.Duration {
	return d.Pong.Received.Sub(d.Ping.Sent)
}
