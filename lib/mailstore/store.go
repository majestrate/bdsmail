package mailstore

import (
	"io"
)

type Store interface {
	Ensure() error
	Deliver(io.Reader) (Message, error)
	ListNew() ([]Message, error)
	// process a new message, mark it as no longer new and return the message after being processed
	Process(msg Message) (Message, error)
	// List all non-new messages
	List() ([]Message, error)
}
