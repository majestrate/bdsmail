package mailstore

type Message interface {
	Filepath() string
	Filename() string
	Remove() error
}
