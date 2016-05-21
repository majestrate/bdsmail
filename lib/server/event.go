package server

import (
	"bytes"
	"net"
)

// event fired when we got a new mail message
type MailEvent struct {
	// remote address of sender
	Addr net.Addr
	// recipiant of message
	Recip string
	// sender of message
	Sender string
	// body of message
	Body *bytes.Buffer
}

func (ev *MailEvent) Read(d []byte) (int, error) {
	return ev.Body.Read(d)
}
