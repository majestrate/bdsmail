package sendmail

import (
	"github.com/majestrate/bdsmail/lib/mailstore"
	log "github.com/sirupsen/logrus"
	"os"
)

type LocalDeliverJob struct {
	st     mailstore.Store
	result chan bool
	fpath  string
}

// new local delivery job
func NewLocalDelivery(st mailstore.Store, fpath string) DeliverJob {
	return &LocalDeliverJob{
		st:     st,
		result: make(chan bool),
		fpath:  fpath,
	}
}

// local delivery is not cancelable
// TODO: make this configurable
func (l *LocalDeliverJob) Cancel() {
}

// wait for completion
func (l *LocalDeliverJob) Wait() bool {
	return <-l.result
}

// run local delivery
func (l *LocalDeliverJob) Run() {
	var msg mailstore.Message
	f, err := os.Open(l.fpath)
	if err == nil {
		msg, err = l.st.Deliver(f)
		f.Close()
	}
	if err != nil {
		log.Warnf("local delivery failed: %s", err.Error())
		l.result <- false
	}
	// inform result
	l.result <- msg != nil
}
