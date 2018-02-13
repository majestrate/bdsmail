package pop3

import (
	"bds/lib/mailstore"
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
)

// function that authenticates a user
type UserAuthenticator func(string, string) (bool, error)

// pop3 server
type Server struct {
	// obtains a mail store given a user
	Local mailstore.MailRouter
	// login authenticator
	Auth UserAuthenticator
	// server name
	name string
	// tls config
	TLS *tls.Config
}

func New() *Server {
	host, _ := os.Hostname()
	return &Server{
		name: host,
	}
}

// get all messages in mail store
func (s *Server) getMessages(user string) (msgs []mailstore.Message, err error) {
	// get user's mail store
	var st mailstore.Store
	if s.Local == nil {
		err = errors.New("could't find mail store")
		return
	}
	var has bool
	st, has = s.Local.FindStoreFor(user)
	if !has {
		err = errors.New("no such local user")
		return
	}
	var ms []mailstore.Message
	ms, err = st.ListNew()
	// move new mail into cur	msgs
	if err == nil {
		for _, m := range ms {
			_, err = st.Process(m)
			if err != nil {
				log.Errorf("error processing maildir: %s", err.Error())
			}
		}
	}
	// get list of messages
	ms, err = st.List()
	if err == nil {
		for _, msg := range ms {
			msgs = append(msgs, msg)
		}
	}
	return
}

func (s *Server) checkUser(user, passwd string) (allowed bool) {
	if s.Auth != nil {
		allowed, _ = s.Auth(user, passwd)
	}
	return
}

// get all messages and octet count
func (s *Server) obtainMessages(user string) (msgs []mailstore.Message, o int64, err error) {
	msgs, err = s.getMessages(user)
	if err == nil {
		for _, msg := range msgs {
			var info os.FileInfo
			info, err = os.Stat(msg.Filepath())
			if err == nil && !info.IsDir() {
				o += info.Size()
			}
		}
	} else {
		log.Errorf("pop3: %s", err.Error())
	}
	return
}

// pop3 session handler
type pop3Session struct {
	// network connection
	c *textproto.Conn
	// parent server
	s *Server
	// are we in transaction state?
	transaction bool
	// current user
	user string
	// messages we have in this transaction
	msgs []mailstore.Message
	// how many octets we have for all messages
	octs int64
	// messages to delete
	dels []mailstore.Message
}

// run pop3 session mainloop
func (p *pop3Session) Run() {
	// send banner
	err := p.OK("POP3 Server Ready")
	for err == nil {
		var line string
		line, err = p.c.ReadLine()
		if err == nil {
			if strings.ToUpper(line) == "QUIT" {
				// check for quit command
				err = p.OK("k bai")
				break
			} else if p.transaction {
				err = p.handleTransactionLine(line)
			} else {
				err = p.handleLine(line)
			}
		}
	}
	if err != nil && err != io.EOF {
		log.Errorf("error in pop3 session: %s", err.Error())
	}
	// close connection
	p.c.Close()
	// delete old messages
	for _, msg := range p.dels {
		os.Remove(msg.Filepath())
	}
}

// handle line when in transaction mode
func (p *pop3Session) handleTransactionLine(line string) (err error) {
	var info os.FileInfo
	var idx int
	parts := strings.Split(line, " ")
	cmd := strings.ToUpper(parts[0])
	switch cmd {
	case "DELE":
		if len(parts) == 2 {
			idx, err = strconv.Atoi(parts[1])
			if err == nil && (idx > 0 && idx <= len(p.msgs)) {
				// valid, add it to delete
				p.dels = append(p.dels, p.msgs[idx-1])
				p.OK("")
			} else {
				// invalid
				err = p.Error(err.Error())
			}
		}
	case "RETR":
		if len(parts) == 2 {
			idx, err = strconv.Atoi(parts[1])
			if err == nil && (idx > 0 && idx <= len(p.msgs)) {
				// valid
				msg := p.msgs[idx-1].Filepath()
				var f *os.File
				f, err = os.Open(msg)
				if err == nil {
					r := bufio.NewReader(f)
					info, err = f.Stat()
					if err == nil {
						err = p.c.PrintfLine("+OK %d octets", info.Size())
						for err == nil {
							// send line
							line, err = r.ReadString(10)
							if err == io.EOF {
								err = nil
								break
							} else if err == nil {
								line = strings.Trim(line, "\r")
								line = strings.Trim(line, "\n")
								if line == "." {
									line = " ."
								}
								// send line
								err = p.c.PrintfLine(line)
							} else {
								// error
								break
							}
						}
					}
					f.Close()
					// end
					err = p.c.PrintfLine(".")
				} else {
					err = p.Error(err.Error())
				}
			} else {
				// invalid
				err = p.Error("bad message")
			}
		} else {
			err = p.Error("invalid syntax")
		}
		break
	case "UIDL":
		if len(parts) == 2 {
			// 1 message
			idx, err = strconv.Atoi(parts[1])
			if err == nil && (idx > 0 && idx <= len(p.msgs)) {
				// valid
				err = p.OK(parts[1] + " " + p.msgs[idx-1].Filename())
			} else {
				// invalid
				err = p.Error("bad message")
			}
		} else {
			// all messages
			p.OK("")
			dw := p.c.DotWriter()
			for idx, msg := range p.msgs {
				fmt.Fprintf(dw, "%d %s\r\n", 1+idx, msg.Filename())
			}
			// FLUSH :D
			err = dw.Close()
		}
		// begin
		break
	case "STAT":
		// begin
		_, err = fmt.Fprintf(p.c.W, "+OK %d %d", len(p.msgs), p.octs)

		if err == nil {
			// done writing list
			err = p.c.PrintfLine("")
		} else {
			err = p.Error(err.Error())
		}
		break
	case "LIST":
		if len(parts) == 2 {
			// 1 message
			idx, err = strconv.Atoi(parts[1])
			if err == nil {
				if idx > 0 && idx <= len(p.msgs) {
					info, err = os.Stat(p.msgs[idx-1].Filepath())
					if err == nil {
						err = p.c.PrintfLine("+OK %d %d", idx, info.Size())
					}
				} else {
					// no existing message
					err = p.Error("no such message")
				}
			}
		} else {
			// all messages
			err = p.c.PrintfLine("+OK %d messages (%d octets)", len(p.msgs), p.octs)
			if err == nil {
				dw := p.c.DotWriter()
				for i, msg := range p.msgs {
					info, err = os.Stat(msg.Filepath())
					if err == nil {
						// write out entry
						fmt.Fprintf(dw, "%d %d\r\n", i+1, info.Size())
					} else {
						log.Errorf("error in pop3, stat(): %s", err.Error())
					}
				}
				// FLUSH IT !!! :-DDDDD
				err = dw.Close()
			}
		}
		break
	default:
		err = p.Error("bad command")
	}
	return
}

// handle 1 line of input when not in transaction mode
func (p *pop3Session) handleLine(line string) (err error) {
	parts := strings.Split(line, " ")
	cmd := strings.ToUpper(parts[0])
	switch cmd {
	case "NOOP":
		err = p.OK("")
		break
	case "PASS":
		if p.s.checkUser(p.user, line[5:]) {
			p.msgs, p.octs, err = p.s.obtainMessages(p.user)
			if err == nil {
				err = p.c.PrintfLine("+OK %s maildrop logged in, you have %d messages (%d octets)", p.user, len(p.msgs), p.octs)
				p.transaction = err == nil
			} else {
				err = p.Error(err.Error())
			}
		} else {
			err = p.Error("bad login")
		}
		break
	case "USER":
		if len(parts) > 1 {
			p.user = line[5:]
		}
		err = p.OK("you may try login as " + p.user)
		break
	default:
		err = p.Error("bad command")
	}
	return
}

func (p *pop3Session) OK(msg string) (err error) {
	if len(msg) == 0 {
		err = p.c.PrintfLine("+OK")
	} else {
		err = p.c.PrintfLine("+OK %s", msg)
	}
	return
}

// send error
func (p *pop3Session) Error(msg string) (err error) {
	err = p.c.PrintfLine("-ERR %s", msg)
	return
}

// serve sessions with connections accepted from a net.Listener
func (s *Server) Serve(l net.Listener) (err error) {
	for err == nil {
		var c net.Conn
		c, err = l.Accept()
		if err == nil {
			p := &pop3Session{
				c: textproto.NewConn(c),
				s: s,
			}
			go p.Run()
		}
	}
	return
}
