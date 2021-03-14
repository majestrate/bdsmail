package maildir

import (
	"os"
	"path/filepath"
	"strings"
)

type Message string

func (m Message) Filepath() string {
	return string(m)
}

func (m Message) Filename() (f string) {
	_, f = filepath.Split(m.Filepath())
	return
}

func (m Message) Remove() error {
	return os.Remove(m.Filepath())
}

func (m Message) Name() string {
	return strings.Split(m.Filename(), ":")[0]
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
