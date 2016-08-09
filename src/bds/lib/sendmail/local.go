package sendmail

import (
	log "github.com/Sirupsen/logrus"
	"bds/lib/maildir"
	"os"
)


type LocalDeliverJob struct {
	mailDir maildir.MailDir
	result chan bool
	fpath string
}

// local delivery is not cancelable
// TODO: make this configurable
func (l *LocalDeliverJob) Cancel() {
}

// wait for completion
func (l *LocalDeliverJob) Wait() bool {
	return <- l.result
}

// run local delivery
func (l *LocalDeliverJob) Run() {
	var msg maildir.Message
	f, err := os.Open(l.fpath)
	if err == nil {
		msg, err = l.mailDir.Deliver(f)
		f.Close()
	}
	if err != nil {
		log.Warnf("local delivery failed: %s", err.Error())
		l.result <- false
	}
	// inform result
	l.result <- msg != ""
}
