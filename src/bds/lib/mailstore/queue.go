package mailstore

type SendQueue interface {
	Ensure() error
	Offer(msg Message) error
	Pop() (Message, bool)
}
