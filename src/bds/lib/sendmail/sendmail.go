package sendmail

import (
	"bds/lib/mailstore"
	"bds/lib/smtp"
	"errors"
	log "github.com/Sirupsen/logrus"
	"net"
	"strings"
	"sync"
)

var ErrNoLocalMailDelivery = errors.New("no local mail store for user")

// mail bounce handler
// paramters are (recipiant email address, from email address, the filepath of the message, network related error or nil for regular bounce)
type Bouncer func(string, string, string, error)

type connection struct {
	addr   string
	cl     *smtp.Client
	access sync.RWMutex
}

// safely visit a conection with a function
func (c *connection) visit(visitor func(*smtp.Client) error) (err error) {
	c.access.Lock()
	err = visitor(c.cl)
	c.access.Unlock()
	return
}

func (c *connection) Quit() {
	c.cl.Quit()
}

// mail delivery
type Mailer struct {
	// get mail storage for local user
	Local mailstore.MailRouter
	// a dial function to obtain outbound smtp client
	Dial Dialer
	// number of times to try to deliver mail
	// 0 for unlimited
	Retries int
	// called if mail bounces or fails to deliver
	Bounce Bouncer
	// domain resolver function
	Resolve Resolver
	// delivery success hook, called with (recipiant email address, from email address)
	Success func(string, string)
	// for pipelining
	conns map[string]*connection
	// mutex for conns
	cmtx sync.RWMutex
}

// create a new pooled mailer
func NewMailer() *Mailer {
	return &Mailer{
		conns: make(map[string]*connection),
	}
}

// call a visitor for each connection in connection poll
func (s *Mailer) foreach(visitor func(*smtp.Client) error) {
	var conns []*connection
	s.cmtx.Lock()
	for _, conn := range s.conns {
		conns = append(conns, conn)
	}
	s.cmtx.Unlock()
	for _, conn := range conns {
		// in paralell
		go func(c *connection, v func(*smtp.Client) error) {
			err := c.visit(v)
			if err != nil {
				// remove connection on error
				s.delConn(c.addr)
			}
		}(conn, visitor)
	}
}

// delete connection from pool
func (s *Mailer) delConn(addr string) {
	s.cmtx.Lock()
	// safe delete
	c, ok := s.conns[addr]
	if ok {
		c.Quit()
		delete(s.conns, addr)
	}
	s.cmtx.Unlock()
}

// visit a pooled connection in a safe way
func (s *Mailer) visitConn(n, addr, localname string, dialer Dialer, visitor func(*smtp.Client) error) (err error) {
	var c *connection
	c, err = s.getConn(n, addr, localname, dialer)
	if err == nil {
		err = c.visit(visitor)
		if err != nil {
			log.Errorf("failed to visit connection: %s", err.Error())
			s.delConn(addr)
		}
	}
	return
}

// get connection from pool, create if not there
func (s *Mailer) getConn(n, addr, localname string, dial Dialer) (sc *connection, err error) {
	s.cmtx.Lock()
	defer s.cmtx.Unlock()
	var ok bool
	sc, ok = s.conns[addr]
	if !ok {
		var c net.Conn
		c, err = dial(n, addr)
		if err == nil {
			// new connection
			var cl *smtp.Client
			cl, err = smtp.NewClient(c, addr)
			if err != nil {
				log.Errorf("failed to dial: %s", err.Error())
				return
			}
			err = cl.Hello(localname)
			if err != nil {
				log.Errorf("failed to helo: %s", err.Error())
				cl.Quit()
				return
			}
			sc = &connection{
				cl:   cl,
				addr: addr,
			}
			// success
			s.conns[addr] = sc
		}
	}
	return
}

// try delivering mail
// returns a DeliveryJob that can be cancelled
func (s *Mailer) Deliver(recip, from string, msg mailstore.Message) (d DeliverJob) {
	log.Infof("Delivering %s to %s from %s", msg.Filepath(), recip, from)
	dialer := s.Dial
	if dialer == nil {
		dialer = net.Dial
	}

	bounce := s.Bounce

	resolver := s.Resolve

	if resolver == nil {
		resolver = func(name string) (a net.Addr, err error) {
			log.Debugf("mx lookup for %s", name)
			var mx []*net.MX
			mx, err = net.LookupMX(name)
			if err != nil && mx != nil {
				for _, m := range mx {
					var ips []net.IP
					log.Debugf("resolve mx record %s", m.Host)
					ips, err = net.LookupIP(m.Host)
					if err == nil {
						for _, ip := range ips {
							a, err = net.ResolveIPAddr("ip", ip.String())
							if err == nil {
								log.Debugf("resolved %s to %s", name, a)
								return
							}
						}
					}
				}
			}
			return
		}
	}
	var st mailstore.Store
	if s.Local != nil {
		st, _ = s.Local.FindStoreFor(recip)
	}

	if st == nil {
		d = &RemoteDeliverJob{
			unlimited: s.Retries == 0,
			cancel:    false,
			retries:   s.Retries,
			visit: func(f func(*smtp.Client) error) error {
				parts := strings.Split(recip, "@")
				if len(parts) == 2 {
					r_addr := parts[1]
					a, err := resolver(r_addr)
					if err == nil {
						err = s.visitConn(a.Network(), a.String(), r_addr, dialer, f)
					} else {
						log.Warnf("failed to resolve %s: %s", r_addr, err.Error())
					}
					return err
				} else {
					log.Warnf("bad email address %s", recip)
				}
				return nil
			},
			bounce:    bounce,
			recip:     recip,
			from:      from,
			fpath:     msg.Filepath(),
			result:    make(chan bool),
			delivered: s.Success,
		}
	} else {
		d = &LocalDeliverJob{
			st:     st,
			result: make(chan bool),
			fpath:  msg.Filepath(),
		}
	}
	return
}

// gracefully quit all polled connections and close down
func (m *Mailer) Quit() {
	log.Info("shutting down pooled mailer")
	m.cmtx.Lock()
	// collect connections
	var conns []*connection
	for _, c := range m.conns {
		conns = append(conns, c)
	}
	// close all connections
	for _, c := range conns {
		c.visit(func(cl *smtp.Client) error {
			return cl.Quit()
		})
	}
	// collect all names
	var names []string
	for n, _ := range m.conns {
		names = append(names, n)
	}
	// delete all connections
	for _, n := range names {
		delete(m.conns, n)
	}
	m.cmtx.Unlock()
}
