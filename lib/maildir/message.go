package maildir

import (
	"strings"
)

type Message string

func (m Message) Filepath() string {
	return string(m)
}

func (m Message) Name() string {
	return strings.Split(string(m), ":")[0]
}

// get flags on this message
func (m Message) GetFlags() (flags []Flag) {
	s := m.Filepath()
	if strings.Count(s, ":2,") == 1 {
		// we have flags
		f := strings.Split(s, ",")[1]
		for _, fl := range f {
			flags = append(flags, Flag(fl))
		}
	}
	return
}
