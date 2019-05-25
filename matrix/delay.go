package matrix

import "time"

// Delay represents a delay for a room.
type Delay struct {
	Room, ID   string
	PingTime, PongTime time.Time
	MatrixTime, ReceiveTime time.Time
}

func (d *Delay) Ping() time.Duration {
	return d.PongTime.Sub(d.PingTime)
}

func (d *Delay) Pong() time.Duration {
	return d.ReceiveTime.Sub(d.PongTime)
}

func (d *Delay) RTT() time.Duration {
	return d.ReceiveTime.Sub(d.PingTime)
}

func (d *Delay) IRCPong() time.Duration {
	return d.MatrixTime.Sub(d.PongTime)
}

func (d *Delay) MatrixPong() time.Duration {
	return d.ReceiveTime.Sub(d.MatrixTime)
}
