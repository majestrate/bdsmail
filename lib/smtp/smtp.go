package smtp

import (
  "bytes"
  "fmt"
  "io"
  "net"
  "net/smtp"
  "net/textproto"
  "regexp"
  "strings"
  "time"
)

var (
	rcptToRE   = regexp.MustCompile(`[Tt][Oo]:<(.+)>`)
	mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:<(.*)>`) // Delivery Status Notifications are sent with "MAIL FROM:<>"
)


// create a new smtp client
// wrapper function
func NewClient(conn net.Conn, host string) (*smtp.Client, error) {
  return smtp.NewClient(conn, host)
}

// smtp message handler
type Handler func(remoteAddr net.Addr, from string, to []string, body []byte)


// serve smtp via a net.Listener
func Serve(l net.Listener, h Handler, appname, hostname string) (err error) {
  serv := Server{
    Appname: appname,
    Hostname: hostname,
    Handler: h,
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
}

type session struct {
  srv *Server
  conn *textproto.Conn
  raddr net.Addr
  laddr net.Addr
  remoteName string
}

func (s *Server) newSession(conn net.Conn) *session {
  return &session{
    srv: s,
    conn: textproto.NewConn(conn),
    raddr: conn.RemoteAddr(),
    laddr: conn.LocalAddr(),
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
      c.PrintfLine("250 %s greets %s", s.srv.Hostname, s.remoteName)
      from = ""
      to = nil
      body.Reset()
    case "MAIL":
      match := mailFromRE.FindStringSubmatch(args)
      if match == nil {
        // no match
        c.PrintfLine("501 syntax error in parameters (invalid FROM)")
      } else {
        from = match[1]
        c.PrintfLine("250 Ok")
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
        break
      }
      match := rcptToRE.FindStringSubmatch(args)
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
    case "DATA":
      if from == "" || to == nil {
        c.PrintfLine("503 bad sequence of commands, (MAIL & RCPT Required befored DATA)")
        break
      }
      // read mail body
      c.PrintfLine("354 Start giving me the mail yo, end with <CR><LF>.<CR><LF>")
      body.Reset()
      // put recvived header
      now := time.Now().Format("Mon, _2 Jan 2006 22:04:05 -0000 (UTC)")
      fmt.Fprintf(&body, "Received: from %s (%s [127.0.0.1])\r\n", s.remoteName, s.raddr)
      fmt.Fprintf(&body, "        by %s (%s) with SMTP\r\n", s.srv.Hostname, s.srv.Appname)
      fmt.Fprintf(&body, "        for <%s>; %s\r\n", to[0], now)
      dr := c.DotReader()
      _, err = io.Copy(&body, dr)
      if err == nil {
        // copy gud
        c.PrintfLine("250 Ok: queued")
        if s.srv.Handler != nil {
          buff := make([]byte, body.Len())
          copy(buff, body.Bytes())
          go s.srv.Handler(s.raddr, from, to, buff)
        }
      } else {
        // something bad happened?
      }
      from = ""
      to = nil
      body.Reset()
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
