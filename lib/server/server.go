package server

import (
	"bytes"
	"github.com/majestrate/botemail/lib/lua"
	"github.com/majestrate/botemail/lib/maildir"
	"github.com/mhale/smtpd"
	log "github.com/Sirupsen/logrus"
	"net"
)

// handler of mail messages
type MailHandler interface {
	// we got a mail message
	// handle it somehow
	GotMail(ev *MailEvent)
	// do we accept mail going to this recipiant?
	AllowRecipiant(recip string) bool
}

// botemail mail server
type Server struct {

	// mail handler
	Handler MailHandler


	// unexported fields
	
	l *lua.Lua
	serv *smtpd.Server
	listener net.Listener
	chnl chan *MailEvent
	mail maildir.MailDir

}

// bind server to address in config
func (s *Server) Bind() (err error) {
	addr, ok := s.l.GetConfigOpt("bind")
	if ! ok {
		addr = ":25"
	}
	log.Info("Bind mail server to", addr)
	s.listener, err = net.Listen("tcp", addr)
	s.serv.Addr = addr
	return
}

// queue mail to be filtered
func (s *Server) queueMail(addr net.Addr, from string, to []string, body []byte) {
	// for each recip fire a mail event
	for _, recip := range to {
		ev := &MailEvent{
			Addr: addr,
			Sender: from,
			Recip: recip,
			Body: bytes.NewBuffer(body),
		}
		s.chnl <- ev
	}
}

// we got mail that was not dropped by the filters
func (s *Server) gotMail(ev *MailEvent) (err error) {
	log.Info("we got mail for ", ev.Recip, " from ", ev.Sender)
	// deliver to maildir if set
	if s.mail.String() != "" {
		r := bytes.NewReader(ev.Body.Bytes())
		// deliver
		err = s.mail.Deliver(r)
	}
	if s.Handler != nil {
		go s.Handler.GotMail(ev)
	}
	return
}

// run a lua filter given a mail event
// return the code returned by the lua function
func (s *Server) runFilter(filtername string, ev *MailEvent) int {
	log.Debug(`running filter "`+filtername+`"...`)
	ret := s.l.CallMailFilter(filtername, ev.Recip, ev.Sender, ev.Body.String())
	log.Debug(`filter "`+filtername+`" returned `, ret)
	return ret
}

// called for each recipiant
// checks mail message against whitelist, blacklist and
// checkspam filters sequentially
func (s *Server) filterMail(ev *MailEvent) (err error) {
	// check whitelist filter
	if s.runFilter("whitelist", ev) == 1 {
		// explicit whitelist
		err = s.gotMail(ev)
		return
	}
	// check blacklist filter
	if s.runFilter("blacklist", ev) == 1 {
		// drop message
		log.WithFields(log.Fields{
			"addr": ev.Addr,
			"recip": ev.Recip,
			"sender": ev.Sender,
			"msgsize": ev.Body.Len(),
		}).Info("message hit blacklist")
		return
	}
	if s.runFilter("checkspam", ev) == 1 {
		// we got a spam message
		log.WithFields(log.Fields{
			"addr": ev.Addr,
			"recip": ev.Recip,
			"sender": ev.Sender,
			"msgsize": ev.Body.Len(),
		}).Info("message hit spam filter")
		return
	}
	// this mail was accepted
	err = s.gotMail(ev)
	return
}

func (s *Server) Run() {
	defer s.end()
	// run acceptor
	go func() {
		log.Info("Serving SMTP server on ", s.listener.Addr())
		err := s.serv.Serve(s.listener)
		if err != nil {
			log.Fatal("smtp fail ", err)
		}
	}()
	log.Debug("run mail")
	for {
		// filtering
		ev , ok := <- s.chnl
		if ! ok {
			log.Debug("exiting mainloop")
			return
		}
		recip := ev.Recip
		if s.allowRecip(recip) {
			err := s.filterMail(ev)
			if err != nil {
				log.Error("Error while handling inbound mail ", err)
			}
		} else {
			log.Info("Ingoring message with invalid recipiant ", recip)
		}
	}
}

// do we allow a recipiant ?
func (s *Server) allowRecip(recip string) (allow bool) {
	if s.Handler == nil {
		// default to always allowing messages
		allow = true
	} else {
		allow = s.Handler.AllowRecipiant(recip)
	}
	return
}


// end serving
func (s *Server) end() {
	s.l.Close()
	s.listener.Close()
	close(s.chnl)
}

// load configuration file
func (s *Server) LoadConfig(fname string) (err error) {
	log.Debug("Load config file ", fname)
	err = s.l.LoadFile(fname)
	if err == nil {
		str, _ := s.l.GetConfigOpt("maildir")
		if len(str) > 0 {
			log.Info("Using maildir at ", str)
			s.mail = maildir.MailDir(str)
			err = s.mail.Ensure()
		}
		if err == nil {
			str, _ := s.l.GetConfigOpt("domain")
			if len(str) == 0 {
				str = "localhost"
			}
			log.Info("Setting mail hostname to ", str)
			s.serv.Hostname = str
		}
	}
	return
}

func New() (s *Server) {
	s = &Server{
		chnl: make(chan *MailEvent, 1024),
		l: lua.New(),
		serv: &smtpd.Server{
			Appname: "botemail",
		},
	}
	if s.l.JIT() != nil {
		log.Fatal("failed to initialize luajit")
	}
	s.serv.Handler = s.queueMail
	return
}
