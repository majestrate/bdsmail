package maildir

import (
	"github.com/majestrate/bdsmail/lib/mailstore"
	"io"
	"os"
)

type mailQueue struct {
	md   MailDir
	last string
}

func (q *mailQueue) Ensure() error {
	return q.md.Ensure()
}

// deliver a new message into the internal message queue
func (q *mailQueue) Offer(msg mailstore.Message) (err error) {
	var f io.ReadCloser
	f, err = os.Open(msg.Filepath())
	if err == nil {
		_, err = q.md.Deliver(f)
		f.Close()
	}
	return
}

// move all new messages into cur pop one from cur
func (q *mailQueue) Pop() (msg mailstore.Message, has bool) {
	// pump all new messages
	msgs, err := q.md.listDir("new")
	if err == nil && len(msgs) > 0 {
		for _, m := range msgs {
			q.md.ProcessNew(m)
		}
	}
	msgs, err = q.md.ListCur()
	if err == nil && len(msgs) > 0 {
		if msgs[0].Filepath() == q.last {
			return
		}
		msg = msgs[0]
		q.last = msg.Filepath()
		has = true
	}
	return
}

func MailQueue(d MailDir) mailstore.SendQueue {
	return &mailQueue{
		md: d,
	}
}
