package server

// event fired when we got a new mail message
type MailEvent struct {
	// remote address of sender
	Addr string
	// recipiant of message
	Recip string
	// sender of message
	Sender string
	// file containg the message
	File string
}
