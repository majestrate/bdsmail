package smtp

import (
	"bds/lib/mailstore"
	mail "bds/lib/mailutil"
	"bds/lib/starttls"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"
)

var (
	rcptToRE   = regexp.MustCompile(`[Tt][Oo]:<(.+)>`)
	mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:<(.*)>`) // Delivery Status Notifications are sent with "MAIL FROM:<>"
)

type Client struct {
	smtp.Client
}

// create a new smtp client
// wrapper function
func NewClient(conn net.Conn, host string) (*Client, error) {
	cl, err := smtp.NewClient(conn, host)
	if err == nil {
		return &Client{*cl}, nil
	}
	return nil, err
}

// smtp message handler
type Handler func(remoteAddr net.Addr, from string, to []string, fpath string)

// serve smtp via a net.Listener
func Serve(l net.Listener, h Handler, appname, hostname string) (err error) {
	serv := Server{
		Appname:  appname,
		Hostname: hostname,
		Handler:  h,
	}
	return serv.Serve(l)
}

type Server struct {
	// name name of the smtp application
	Appname string
	// the hostname of the smtp server
	Hostname string
	// the handler of inbound mail
	Handler Handler
	// mail storage for inbound mail
	Inbound mailstore.Store
	// outbound mail queue
	Outbound mailstore.SendQueue
	// user authenticator for sending mail
	Auth Auth
	// TLS Config
	TLS *tls.Config
}

type session struct {
	srv        *Server
	conn       *textproto.Conn
	nc         net.Conn
	remoteName string
	user       string
}

func (s *Server) newSession(conn net.Conn) *session {
	return &session{
		srv:  s,
		conn: textproto.NewConn(conn),
		nc:   conn,
	}
}

// serve creates a new smtp sesion after a network connection is established
func (s *Server) Serve(l net.Listener) (err error) {
	defer l.Close()
	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			}
			return
		}
		session := s.newSession(conn)
		go session.serve()
	}
	return
}

// parse smtp line
func parseLine(line string) (cmd string, args string) {
	if idx := strings.Index(line, " "); idx > 0 {
		cmd = strings.ToUpper(line[:idx])
		args = strings.TrimSpace(line[idx+1:])
	} else {
		cmd = strings.ToUpper(line)
	}
	return
}

// handles inbound connection
func (s *session) serve() {
	defer s.conn.Close()
	var from string
	var to []string
	var body bytes.Buffer
	c := s.conn
	c.PrintfLine("220 %s %s SMTP is ready", s.srv.Hostname, s.srv.Appname)
	for {
		line, err := c.ReadLine()
		if err != nil {
			break
		}
		cmd, args := parseLine(line)
		switch cmd {
		case "EHLO", "HELO":
			s.remoteName = args
			c.PrintfLine("250-%s Hello %s", s.srv.Hostname, s.remoteName)
			if s.srv.Auth != nil && cmd == "EHLO" {
				c.PrintfLine("250-AUTH PLAIN")
				if s.srv.TLS != nil {
					c.PrintfLine("250-STARTTLS")
				}
			}
			c.PrintfLine("250 HELP")
			from = ""
			to = nil
			body.Reset()
		case "MAIL":
			match := mailFromRE.Copy().FindStringSubmatch(args)
			if match == nil {
				// no match
				c.PrintfLine("501 syntax error in parameters (invalid FROM)")
			} else {
				if s.srv.Auth != nil && !s.srv.Auth.PermitSend(match[1], s.user) {
					c.PrintfLine("450 4.7.1 not authorized to send")
				} else {
					from = match[1]
					c.PrintfLine("250 Ok")
				}
			}
			to = nil
			body.Reset()

		case "RSET":
			c.PrintfLine("250 Ok")
			from = ""
			to = nil
			body.Reset()
		case "RCPT":
			if from == "" {
				c.PrintfLine("503 bad sequence of commands")
			} else {
				match := rcptToRE.Copy().FindStringSubmatch(args)
				if match == nil {
					// no match
					c.PrintfLine("501 syntax error in parameters (invalid TO)")
				} else {
					if len(to) == 100 {
						// too many recipiants
						c.PrintfLine("452 too many recipients")
					} else {
						to = append(to, match[1])
						c.PrintfLine("250 Ok")
					}
				}
			}
		case "DATA":
			if from == "" || to == nil {
				c.PrintfLine("503 bad sequence of commands, (MAIL & RCPT Required befored DATA)")
				break
			}
			// read mail body
			c.PrintfLine("354 Start giving me the mail yo, end with <CR><LF>.<CR><LF>")
			// put recvived header
			err = mail.WriteRecvHeader(&body, to[0], s.remoteName, s.nc.RemoteAddr().String(), s.srv.Hostname, s.srv.Appname)
			dr := c.DotReader()
			// deliver to maildir
			mr := io.MultiReader(&body, dr)
			var msg mailstore.Message
			msg, err = s.srv.Inbound.Deliver(mr)
			if err == nil {
				if s.srv.Handler == nil {
					// no handler
				} else {
					go s.srv.Handler(s.nc.RemoteAddr(), from, to, msg.Filepath())
				}
			}
			if err == nil {
				c.PrintfLine("250 Ok: Delivered")
			} else {
				log.Error("smtp server error: %s", err.Error())
				c.PrintfLine("500 Error delivering message: %s", err.Error())
			}
			from = ""
			to = nil
			body.Reset()
		case "STARTTLS":
			nc, e := s.startTLS()
			if e == nil {
				c = nc
			} else {
				c.Close()
				return
			}
		case "AUTH":
			if s.srv.Auth == nil {
				// XXX: should we always succeed?
				c.PrintfLine("235 2.7.0 Authentication Succeeded")
			} else {
				parts := strings.Split(args, " ")
				if len(parts) > 1 {
					if parts[0] == "PLAIN" {
						s.doPlainAuth(c, parts[1])
					} else {
						s.failLogin()
					}
				} else {
					c.PrintfLine("535 5.7.8 Authentication credentials invalid")
				}
			}
		case "QUIT":
			c.PrintfLine("221 %s %s SMTP Closing transmssion channel", s.srv.Hostname, s.srv.Appname)
			return
		case "NOOP":
			c.PrintfLine("250 Ok")
			break
		case "HELP", "VRFY", "EXPN":
			c.PrintfLine("502 command not implemented")
		default:
			c.PrintfLine("500 Syntax error, command unrecodnized")
		}
	}
}

func (s *session) startTLS() (conn *textproto.Conn, err error) {
	if s.srv.TLS == nil {
		s.conn.PrintfLine("500 No STARTTLS")
		err = starttls.ErrTlsNotSupported
	} else {
		s.conn.PrintfLine("220 Ready to start TLS")
		conn, _, err = starttls.HandleStartTLS(s.nc, s.srv.TLS)
		if err != nil {
			log.Errorf("starttls error: %s", err.Error())
		}
	}
	return
}

func (s *session) doPlainAuth(c *textproto.Conn, str string) {
	decoded, e := base64.StdEncoding.DecodeString(str)
	if e == nil {
		p := bytes.Split(decoded, []byte{0})
		if len(p) == 3 {
			user := string(p[1])
			passwd := string(p[2])
			if s.srv.Auth.Plain(user, passwd) {
				s.user = user
				c.PrintfLine("235 2.7.0 Authentication Succeeded")
			} else {
				s.failLogin()
			}
		} else {
			s.failLogin()
		}
	} else {
		s.failLogin()
	}
}

func (s *session) failLogin() error {
	s.conn.PrintfLine("454 4.7.0 Temporary authentication failure")
	s.conn.Close()
	return nil
}
