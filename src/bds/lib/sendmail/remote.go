package sendmail

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// job for delivering mail remotely
type RemoteDeliverJob struct {
	unlimited bool
	cancel    bool
	retries   int

	bounce    Bouncer
	visit     func(func(*smtp.Client) error) error
	delivered func(string, string)

	recip string
	from  string

	fpath string

	result chan bool
}

func extractAddr(email string) (addr string) {
	if strings.HasSuffix(email, "@") {
		addr = "localhost"
	} else {
		idx_at := strings.Index(email, "@")
		if strings.HasSuffix(email, ".b32.i2p") {
			addr = email[idx_at+1:]
		} else if strings.HasSuffix(email, ".i2p") {
			idx_i2p := strings.LastIndex(email, ".i2p")
			addr = fmt.Sprintf("smtp.%s.i2p", email[idx_at+1:idx_i2p])
		} else {
			addr = email[idx_at+1:]
		}
	}
	addr = strings.Trim(addr, ",= \t\r\n\f\b")
	return

}

// cancel delivery
func (d *RemoteDeliverJob) Cancel() {
	d.cancel = true
}

// wait for completion
func (d *RemoteDeliverJob) Wait() bool {
	return <- d.result 
}

// run delivery
func (d *RemoteDeliverJob) Run() {
	tries := 0
	sec := time.Duration(1)
	var err error
	for d.unlimited || tries < d.retries {
		if d.cancel {
			break
		}
		// try visiting connection with tryDeliver method
		err = d.visit(d.tryDeliver)
		if err == nil {
			// it worked, mail delivered
			if d.delivered != nil {
				// call delivered callback
				d.delivered(d.recip, d.from)
			}
			// inform waiting
			d.result <- true
			return
		} else {
			// failed to deliver
			tries++
			log.Warnf("failed to deliver message to %s from %s: %s", d.recip, d.from, err.Error())
			sec *= 2
			if sec > 1024 {
				sec = 1024
			}
			time.Sleep(sec * time.Second)
		}
	}
	// failed to deliver
	log.Errorf("delivery of message to %s failed", d.recip)
	if d.bounce != nil {
		// call bounce hook as needed
		d.bounce(d.recip, d.from, err)
	}
	// inform waiting
	d.result <- false
}

// try delivery
func (d *RemoteDeliverJob) tryDeliver(cl *smtp.Client) (err error) {
	var f *os.File
	// open file
	f, err = os.Open(d.fpath)
	if err != nil {
		return
	}
	defer f.Close()
	// mail from
	err = cl.Mail(d.from)
	if err != nil {
		return
	}
	// recpt to
	err = cl.Rcpt(d.recip)
	if err != nil {
		return
	}
	// data
	var wr io.WriteCloser
	wr, err = cl.Data()
	if err != nil {
		return
	}
	// write body
	_, err = io.Copy(wr, f)
	if err != nil {
		return
	}
	// ... flush
	err = wr.Close()
	if err != nil {
		return
	}
	// reset
	// err = cl.Reset()
	return
}
