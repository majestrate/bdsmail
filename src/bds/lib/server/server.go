package server

import (
	"bds/lib/config"
	"bds/lib/db"
	"bds/lib/i2p"
	"bds/lib/maildir"
	"bds/lib/mailstore"
	"bds/lib/model"
	"bds/lib/pop3"
	"bds/lib/sendmail"
	"bds/lib/smtp"
	"bds/lib/starttls"
	"bds/lib/util"
	"bds/lib/web"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// handler of mail messages
type MailHandler interface {
	// we got a mail message
	// handle it somehow
	GotMail(ev *MailEvent)
	// do we accept mail going to this recipiant?
	AllowRecipiant(recip string) bool
}

// mail server
type Server struct {

	// custom mail handler
	Handler MailHandler

	// unexported fields

	// lua interpreter core
	conf config.Config

	inserv  *smtp.Server
	outserv *smtp.Server

	// listener for smtp recv server
	maillistener net.Listener
	// listener for smtp send server
	smtplistener net.Listener
	// listener for pop3 server
	poplistener net.Listener
	// stream session with i2p router
	session i2p.Session
	// listener for web server
	weblistener net.Listener
	// recv mail events from handlers
	chnl chan *MailEvent
	// directory holding all users's maildirs
	mail string
	// lock to use to ensure 1 thread accessing lua
	luamtx sync.RWMutex
	// filepath to configuration
	configFname string
	// database access object
	dao db.DB
	// web ui
	webHandler http.Handler
	// mail sender
	mailer *sendmail.Mailer
	// pop3 server
	pop *pop3.Server
	// tls config
	TLS *tls.Config
}

// bind network services
func (s *Server) Bind() (err error) {
	// bind web ui
	addr, ok := s.conf.Get("bindweb")
	if !ok {
		addr = "127.0.0.1:8080"
	}
	log.Infof("binding web ui to %s", addr)
	s.weblistener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	// bind pop3 server
	addr, ok = s.conf.Get("bindpop3")
	if !ok {
		addr = "127.0.0.1:1110"
	}
	log.Infof("binding pop3 server to %s", addr)
	s.poplistener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	// keyfile for i2p destination
	keyfile, ok := s.conf.Get("i2pkeyfile")
	if !ok {
		keyfile = "bdsmail-privkey.dat"
	}
	// address of i2p router
	i2paddr, ok := s.conf.Get("i2paddr")
	if !ok {
		i2paddr = "127.0.0.1:7656"
	}
	name := util.RandStr(5)
	log.Info("Starting up I2P connection... hang tight we'll get there")
	// craete session
	session := i2p.NewSession(name, i2paddr, keyfile, make(map[string]string))
	err = session.Open()
	if err == nil {
		// made session

		// get local smtp address
		addr, ok := s.conf.Get("bindmail")
		if !ok {
			addr = "127.0.0.1:2525"
		}
		// bind smtp server
		s.smtplistener, err = net.Listen("tcp", addr)
		if err == nil {
			// success
			s.maillistener = session
			s.session = session
			log.Infof("We are %s", session.B32())
			s.mailer = sendmail.NewMailer()
			s.mailer.Retries = 10
			s.mailer.Dial = session.Dial
			s.mailer.Resolve = func(addr string) (net.Addr, error) {
				return session.LookupI2P(addr)
			}
			s.mailer.Local = s
			s.mailer.Success = func(recip, from string) {
				log.Infof("Delievered mail to %s from %s", recip, from)
			}
		} else {
			// close session we got an error setting up local smtp listener
			session.Close()
		}
	}
	return
}

// dial out
func (s *Server) dial(net, addr string) (c net.Conn, err error) {
	if strings.HasSuffix(addr, ".i2p") {
		c, err = s.session.Dial(net, addr)
	} else {
		// TODO: over tor?
		// c, err = net.Dial(net, addr)
		err = errors.New("cannot dial outside of i2p")
	}
	return
}

// queue mail to be filtered
func (s *Server) queueMail(addr net.Addr, from string, to []string, fpath string) {
	// for each recip fire a mail event
	for _, recip := range to {
		ev := &MailEvent{
			Addr:   addr,
			Sender: from,
			Recip:  recip,
			File:   fpath,
		}
		s.chnl <- ev
	}
}

// get a local maildir or empty string if it's not local to us
func (s *Server) FindStoreFor(email string) (st mailstore.Store, has bool) {
	email = normalizeEmail(email)
	if strings.Count(email, "@") == 1 {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			addr := parts[1]
			user := parts[0]
			if addr == s.session.B32() && s.dao != nil {
				st, has = s.dao.FindStoreFor(user)
			}
		}
	}
	return
}

// we got mail that was not dropped by the filters
func (s *Server) gotMail(ev *MailEvent) (err error) {
	log.Info("we got mail for ", ev.Recip, " from ", ev.Sender)
	st, has := s.FindStoreFor(ev.Recip)
	if !has {
		st, _ = s.dao.FindStoreFor("postmaster")
	}
	// deliver locally
	j := sendmail.NewLocalDelivery(st, ev.File)
	go j.Run()
	ok := j.Wait()
	if ok && s.Handler != nil {
		s.Handler.GotMail(ev)
	}
	os.Remove(ev.File)
	return
}

// run a lua filter given a mail event
// return the code returned by the lua function
func (s *Server) runFilter(filtername string, ev *MailEvent) int {
	return 0
}

// check that a remote address is valid for the recipiant
// this can block for a bit
func (s *Server) i2pSenderIsValid(addr net.Addr, from string) (valid bool) {
	fromAddr := parseFromI2PAddr(normalizeEmail(from))
	if len(fromAddr) > 0 {
		tries := 16
		for tries > 0 {
			log.Infof("looking up recipiant address %s", fromAddr)
			raddr, err := s.session.LookupI2P(fromAddr)
			if err == nil {
				// lookup worked
				valid = raddr.String() == addr.String()
				break
			} else {
				log.Warnf("could not lookup %s", fromAddr)
				tries--
			}
		}
	}
	return
}

// called for each recipiant
// checks mail message against whitelist, blacklist and
// checkspam filters sequentially
func (s *Server) filterMail(ev *MailEvent) (err error) {
	// check invalid address for i2p
	if !s.i2pSenderIsValid(ev.Addr, ev.Sender) {
		// bad address
		log.Warnf("bad i2p address from %s", ev.Sender)
		return
	}

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
			"addr":   ev.Addr,
			"recip":  ev.Recip,
			"sender": ev.Sender,
		}).Info("message hit blacklist")
		return
	}
	if s.runFilter("checkspam", ev) == 1 {
		// we got a spam message
		log.WithFields(log.Fields{
			"addr":   ev.Addr,
			"recip":  ev.Recip,
			"sender": ev.Sender,
		}).Info("message hit spam filter")
		return
	}
	// this mail was accepted
	err = s.gotMail(ev)
	return
}

func (s *Server) Run() {
	// run recv mail acceptor
	go func() {
		log.Info("Serving Inbound SMTP server")
		err := s.inserv.Serve(s.maillistener)
		log.Info("SMTP Server ended")
		if err != nil {
			log.Fatal("inbound smtp died ", err)
		}
	}()

	// run web ui
	go func() {
		if s.webHandler == nil {
			s.webHandler = http.HandlerFunc(http.NotFound)
		}
		log.Info("Serving Web ui")
		err := http.Serve(s.weblistener, s.webHandler)
		if err != nil {
			log.Fatal("web ui died ", err)
		}
	}()

	// run send mail acceptor
	go func() {
		log.Info("Server Outbound SMTP Server on ", s.smtplistener.Addr())
		err := s.outserv.Serve(s.smtplistener)
		if err != nil {
			log.Fatal("outbound smtp died ", err)
		}
	}()

	// run outbound mail flusher
	go func() {
		log.Info("Outbound mail flusher started")
		for s.mailer != nil {
			// flush outbound messages
			s.flushOutboundMailQueue()
			time.Sleep(time.Second * 10)
		}
		log.Info("Outbound mail flusher exited")
	}()

	// run pop3 server
	go func() {
		if s.dao != nil {
			s.pop.Auth = s.dao.CheckUserLogin
			s.pop.Local = s.dao
		}
		log.Info("Serving POP3 server")
		err := s.pop.Serve(s.poplistener)
		if err != nil {
			log.Fatalf("POP3 server died: %s", err.Error())
		}
	}()

	log.Debug("run mail")
	for {
		// filtering
		ev, ok := <-s.chnl
		if !ok {
			log.Info("exiting mainloop")
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
		// allow recip that only match the hostname of the server or the base32 address of the server
		addr := parseFromI2PAddr(recip)
		allow = addr == s.session.B32() || addr == s.inserv.Hostname
	} else {
		// custom mail handler
		allow = s.Handler.AllowRecipiant(recip)
	}
	return
}

// send all pending outbound messages
func (s *Server) flushOutboundMailQueue() {
	log.Debug("flush outbound messages")
	for {
		msg, has := s.outserv.Outbound.Pop()
		if !has {
			return
		}
		log.Infof("sending %s", msg.Filename())
		f, err := os.Open(msg.Filepath())
		if err == nil {
			c := textproto.NewConn(f)
			hdr, err := c.ReadMIMEHeader()
			if err == nil {
				var to []string
				var from string
				from = hdr.Get("From")
				for _, h := range []string{"To", "Cc"} {
					vs, ok := hdr[h]
					if ok {
						to = append(to, vs...)
					}
				}
				c.Close()
				go s.sendOutboundMessage(from, to, msg)
			} else {
				log.Errorf("bad outboud message %s: %s", msg.Filepath(), err.Error())
				c.Close()
			}
		} else {
			log.Errorf("failed to open outbound message %s: %s", msg.Filepath(), err.Error())
		}
	}
}

// send 1 outbound message
func (s *Server) sendOutboundMessage(from string, to []string, msg mailstore.Message) {
	log.Infof("Sending outbound mail %s", msg.Filename())
	var recips []string
	servers := make(map[string]int)
	for _, r := range to {
		r = normalizeEmail(r)
		if len(r) > 0 {
			recips = append(recips, r)
			parts := strings.Split(r, "@")
			server := parts[1]
			// find remove servers
			if server != s.session.B32() {
				num, ok := servers[server]
				if ok {
					servers[server] = num + 1
				} else {
					servers[server] = 1
				}
			}
		}
	}

	if len(recips) == 0 {
		log.Warnf("%s not deliverable, no valid recipiants", msg.Filepath())
		msg.Remove()
		return
	}

	var jobs []sendmail.DeliverJob

	// deliver to all
	for _, recip := range recips {
		// fire off delivery job
		d := s.mailer.Deliver(recip, from, msg)
		jobs = append(jobs, d)
		go d.Run()
	}

	for _, j := range jobs {
		j.Wait()
	}

	msg.Remove()

	return
}

func (s *Server) PermitSend(from, username string) bool {
	user, server := splitEmail(from)
	return username == user && (server == s.inserv.Hostname || server == s.session.B32())
}

func (s *Server) Plain(username, password string) bool {
	good, _ := s.dao.CheckUserLogin(username, password)
	return good
}

// handle mail for sending from inet to i2p
func (s *Server) handleInetMail(remote net.Addr, from string, to []string, fpath string) {
	log.Debugf("handle send mail from %s", remote)
	us := parseFromI2PAddr(from)
	if us == s.inserv.Hostname || us == s.session.B32() {
		// accepted for outbound mail
		log.Infof("outbound message queued: %s", fpath)
	} else {
		log.Errorf("bad outbound mail from %s", us)
		// remove file
		os.Remove(fpath)
	}
}

// stop server
func (s *Server) Stop() {
	close(s.chnl)
	if s.mailer != nil {
		s.mailer.Quit()
		s.mailer = nil
	}
	if s.maillistener != nil {
		s.maillistener.Close()
		s.maillistener = nil
	}
	if s.smtplistener != nil {
		s.smtplistener.Close()
		s.smtplistener = nil
	}
	if s.weblistener != nil {
		s.weblistener.Close()
		s.weblistener = nil
	}
	log.Info("Server Stopped")
}

// load configuration file
func (s *Server) LoadConfig(fname string) (err error) {
	log.Debug("Load config file ", fname)
	s.configFname = fname
	err = s.ReloadConfig()
	return
}

// reload server configuration
func (s *Server) ReloadConfig() (err error) {
	err = s.conf.Load(s.configFname)
	if err != nil {
		return
	}
	str, _ := s.conf.Get("maildir")
	if len(str) == 0 {
		str = "mail"
	}

	str, _ = filepath.Abs(str)
	log.Info("Using user maildir at ", str)
	_, err = os.Stat(str)
	if os.IsNotExist(err) {
		err = os.Mkdir(str, 0700)
	}
	if err != nil {
		return
	}

	s.mail = str

	str, _ = s.conf.Get("inbound_maildir")
	if len(str) == 0 {
		str = "inbound"
	}
	str, _ = filepath.Abs(str)
	log.Info("Using inbound maildir at ", str)
	s.inserv.Inbound = maildir.MailDir(str)
	err = s.inserv.Inbound.Ensure()
	if err != nil {
		return
	}
	// set pop3 server maildir getter
	s.pop.Local = s.dao

	str, _ = s.conf.Get("outbound_maildir")
	if len(str) == 0 {
		str = "outbound"
	}
	str, _ = filepath.Abs(str)
	log.Info("using outbound mail in ", str)
	s.outserv.Auth = s
	s.outserv.Inbound = maildir.MailDir(str)
	s.outserv.Outbound = maildir.MailQueue(maildir.MailDir(str))
	err = s.outserv.Outbound.Ensure()
	if err != nil {
		return
	}
	err = s.outserv.Inbound.Ensure()
	if err != nil {
		return
	}

	str, _ = s.conf.Get("domain")
	if len(str) == 0 {
		if s.session != nil {
			str = s.session.B32()
		} else {
			str = "localhost"
		}
	}
	domain := str
	log.Info("Setting mail hostname to ", domain)
	s.inserv.Hostname = domain
	s.outserv.Hostname = domain

	tkey, _ := s.conf.Get("tls_keyfile")
	if len(tkey) == 0 {
		tkey = "tls-privkey.pem"
	}

	tcert, _ := s.conf.Get("tls_cert")
	if len(tcert) == 0 {
		tcert = "tls-cert.pem"
	}

	log.Info("Ensuring TLS key and certs...")
	s.TLS, err = starttls.GenTLS(domain, "bdsmail", tcert, tkey, 2048)
	if err != nil {
		log.Errorf("failed to generate tls key/cert: %s", err.Error())
		return
	}
	s.outserv.TLS = s.TLS
	s.pop.TLS = s.TLS

	// only initialize dao if not initialized
	if s.dao == nil {
		str, _ = s.conf.Get("database")
		if len(str) == 0 {
			str = "mailserver.sqlite"
		}
		log.Info("Initialize database ", str)
		var dao db.DB
		dao, err = db.NewDB(str)
		if err == nil {
			// run database mainloop
			go dao.Run()
			err = dao.Ensure()
			if err == nil {
				// ensure regular utility users
				for _, name := range []string{"postmaster", "admin", "abuse"} {
					err = dao.EnsureUser(name, func(u *model.User) error {
						u.MailDirPath = filepath.Join(s.mail, name)
						return nil
					})
					if err != nil {
						// failed to ensure this user
						log.Errorf("failed to ensure standard user %s: %s", name, err.Error())
						return
					}
				}

				// ensure all users' resources are there
				err = dao.VisitAllUsers(func(u *model.User) error {
					return u.Ensure()
				})

				if err == nil {
					log.Info("Database ready")
					s.dao = dao
				} else {
					log.Errorf("failed to ensure users: %s", err.Error())
				}
			} else {
				log.Error("database ensure failed: %s", err.Error())
			}
		}
	}
	assetsdir, ok := s.conf.Get("assets")
	if ok && s.dao != nil {
		s.webHandler = web.NewMiddleware(assetsdir, s.dao)
	}
	return
}

// create new server with defaults
func New() (s *Server) {
	Appname := fmt.Sprintf("BDSMail-%s", Version())
	s = &Server{
		chnl: make(chan *MailEvent, 10),
		inserv: &smtp.Server{
			Appname: Appname,
		},
		outserv: &smtp.Server{
			Appname: Appname,
		},
		pop: pop3.New(),
	}
	s.inserv.Handler = s.queueMail
	s.outserv.Handler = s.handleInetMail
	return
}
