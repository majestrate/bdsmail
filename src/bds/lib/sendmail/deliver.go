package sendmail

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"time"
)

// job for delivering mail
type DeliverJob struct {
	unlimited bool
	cancel bool
	retries int
	
	bounce Bouncer
	visit func(func(*smtp.Client) error) error
	delivered func(string, string)
	
	recip string
	from string

	body []byte

	Result chan bool
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
func (d *DeliverJob) Cancel() {
	d.cancel = true
}

// run delivery
func (d *DeliverJob) run() {
	tries := 0
	sec := time.Duration(1)
	var err error
	for (d.unlimited || tries < d.retries) && !d.cancel {
		// try visiting connection with tryDeliver method
		err = d.visit(d.tryDeliver)
		if err == nil {
			// it worked, mail delivered
			if d.delivered != nil {
				// call delivered callback
				d.delivered(d.recip, d.from)
			}
			// inform waiting
			d.Result <- true
			return
		} else {
			// failed to deliver
			tries ++
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
	d.Result <- false
}

// try delivery
func (d *DeliverJob) tryDeliver(cl *smtp.Client) (err error) {
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
	// ... full write
	n := 0
	for n < len(d.body) && err == nil {
		n, err = wr.Write(d.body[n:])
	}
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
