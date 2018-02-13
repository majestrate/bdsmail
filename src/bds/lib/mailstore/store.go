package mailstore

import (
	"io"
)

type Store interface {
	Ensure() error
	Deliver(io.Reader) (Message, error)
	ListNew() ([]Message, error)
}
