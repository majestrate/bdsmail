package i2p

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// implements net.Conn
type samConn struct {
	c net.Conn
	l I2PAddr
	r I2PAddr
}

// implements net.Conn
func (sc *samConn) Close() (err error) {
	err = sc.c.Close()
	return
}

// implements net.Conn
func (sc *samConn) LocalAddr() net.Addr {
	return sc.l
}

// implements net.Conn
func (sc *samConn) RemoteAddr() net.Addr {
	return sc.r
}

// implements net.Conn
func (sc *samConn) Read(d []byte) (n int, err error) {
	n, err = sc.c.Read(d)
	return
}

// implements net.Conn
func (sc *samConn) SetDeadline(t time.Time) (err error) {
	err = sc.c.SetDeadline(t)
	return
}

// implements net.Conn
func (sc *samConn) SetReadDeadline(t time.Time) (err error) {
	err = sc.c.SetReadDeadline(t)
	return
}

// implements net.Conn
func (sc *samConn) SetWriteDeadline(t time.Time) (err error) {
	err = sc.c.SetWriteDeadline(t)
	return
}

// implements net.Conn
func (sc *samConn) Write(d []byte) (n int, err error) {
	n, err = sc.c.Write(d)
	return
}

type samPacketConn struct {
	// sam udp addr
	a *net.UDPAddr
	// local udp conn
	c net.PacketConn
	// parent session
	s *samSession
}

func (s *samPacketConn) LookupI2P(name string) (I2PAddr, error) {
	return s.s.LookupI2P(name)
}

func (s *samPacketConn) Close() (err error) {
	s.s.Close()
	s.c.Close()
	return
}

func (s *samPacketConn) SetDeadline(t time.Time) (err error) {
	err = s.c.SetDeadline(t)
	return
}

func (s *samPacketConn) SetWriteDeadline(t time.Time) (err error) {
	err = s.c.SetWriteDeadline(t)
	return
}

func (s *samPacketConn) SetReadDeadline(t time.Time) (err error) {
	err = s.c.SetReadDeadline(t)
	return
}

func (s *samPacketConn) LocalAddr() net.Addr {
	return s.s.k.addr
}

func (s *samPacketConn) EnsureKeyfile(fname string) (err error) {
	return
}

func (s *samPacketConn) WriteTo(d []byte, to net.Addr) (n int, err error) {
	if to.Network() != "I2P" {
		err = errors.New("cannot send to non i2p network")
		return
	}
	// build packet
	// format is <base64address>\n<payload>
	a := to.String()
	l := len(a)
	sd := make([]byte, len(d)+1+l)
	copy(sd[:], []byte(a))
	sd[l] = 10
	copy(sd[1+l:], d)
	// send it to router
	n, err = s.c.WriteTo(sd, s.a)
	if err == nil {
		n -= 1
		n -= l
	} else {
		n = 0
	}
	return
}

func (s *samPacketConn) ReadFrom(d []byte) (n int, addr net.Addr, err error) {
	var b [65536]byte
	rn, ra, err := s.c.ReadFrom(b[:])
	if err == nil {
		if ra != s.a {
			// bad source address (do something?)
			// sometimes java i2p sends via loopback so eh whatever let's ignore
		}
		// correct address
		i := bytes.Index(b[:], []byte{10})
		if i > 1 {
			addr = I2PAddr(string(b[:i+1]))
			d = b[i+1 : rn]
		}
	}
	return
}

func (sp *samPacketConn) B32() string {
	return sp.s.k.addr.Base32()
}

// keypair
type samKeys struct {
	priv string
	addr I2PAddr
}

// implements i2p.Session i2p.StreamSession and i2p.PacketSession
type samSession struct {
	// name of the sam session
	name string
	// private keys
	k *samKeys
	// control connection
	c net.Conn
	// access mutex for control conntection
	cm sync.RWMutex
	// packet connection
	p net.PacketConn
	// minimum version
	minv string
	// maximum version
	maxv string
}

func (s *samSession) packet(samaddr *net.UDPAddr, bindip net.IP) (p *samPacketConn, err error) {
	var ua *net.UDPAddr
	// resolve local address to bind udp socket to
	// bind to unused udp port on ip6
	ua, err = net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:0", bindip.To16()))
	if err == nil {
		p = &samPacketConn{
			a: samaddr,
			s: s,
		}
		p.c, err = net.ListenUDP(ua.Network(), ua)
		if err == nil {
			// bound
			la := p.c.LocalAddr()
			var host, port string
			host, port, err = net.SplitHostPort(la.String())
			if err == nil {
				// talk to router and establish udp forwarding
				_, err = fmt.Fprintf(s.c, "SESSION CREATE STYLE=DATAGRAM ID=%s DESTINATION=%s HOST=%s PORT=%s\n", s.name, s.k.priv, host, port)
				// read response from udp forward
				r := bufio.NewReader(s.c)
				var line string
				line, err = r.ReadString(10)
				if err == nil {
					if strings.HasPrefix(line, "SESSION STATUS RESULT=OK DESTINATION=") {
						// we gud
					} else {
						// could not create session
						p.c.Close()
						p = nil
						err = errors.New(line)
					}
				} else {
					// did not read reply
					p.c.Close()
					p = nil
				}
			} else {
				// faileed to parse host/port ?! wat.
				p.c.Close()
				p = nil
			}
		} else {
			// error binding?
			p = nil
		}
	}
	return
}

// implements net.PacketConn and net.Listener
func (s *samSession) LocalAddr() net.Addr {
	return s.k.addr
}

// make new connection
func (s *samSession) connect(a string) (c net.Conn, err error) {
	if a == "" {
		a = s.c.RemoteAddr().String()
	}
	c, err = net.Dial("tcp", a)
	// do handshake
	if err == nil {
		// send hello
		_, err = fmt.Fprintf(c, "HELLO VERSION MIN=%s MAX=%s\n", s.minv, s.maxv)
		if err == nil {
			r := bufio.NewReader(c)
			var line string
			// read hello reply
			line, err = r.ReadString(10)
			if err == nil {
				// okay we got a reply
				if line == "HELO REPLY RESULT=NOVERSION" {
					// bad router version
					err = errors.New("router does not support sam version 3")
				} else if strings.HasPrefix(line, "HELLO REPLY RESULT=OK") {
					// yeah sure we got a good reply
					// all is well
					return
				} else {
					// bad line
					err = errors.New(line)
				}
			}
		}
		c.Close()
		c = nil
	}
	return
}

// implements i2p.Session
func (s *samSession) EnsureKeyfile(fname string) (err error) {
	_, err = os.Stat(fname)
	if os.IsNotExist(err) {
		// create keys
		var c net.Conn
		// connect to router
		c, err = s.connect("")
		if err == nil {
			// connected
			// generate destination keys
			_, err = fmt.Fprintf(c, "DEST GENERATE\n")
			if err == nil {
				// read result
				r := bufio.NewReader(c)
				var line string
				line, err = r.ReadString(10)
				sc := bufio.NewScanner(strings.NewReader(line))
				sc.Split(bufio.ScanWords)
				k := new(samKeys)
				// parse result
				for sc.Scan() {
					t := sc.Text()
					if t == "DEST" || t == "REPLY" {
						// control compontent
						continue
					} else if strings.HasPrefix(t, "PUB=") {
						// public compontent
						k.addr = I2PAddr(t[4:])
					} else if strings.HasPrefix(t, "PRIV=") {
						// private component
						k.priv = string(t[5:])
					} else {
						// error
						k = nil
						err = errors.New("failed to parse generated keys: " + t)
					}
				}
				if err == nil {
					// keys were made
					var f io.WriteCloser
					// save key
					f, err = os.Create(fname)
					if err == nil {
						// TODO: use standard format
						_, err = fmt.Fprintf(f, "%s\n%s\n", k.addr, k.priv)
						f.Close()
					}
					// clear keys
				}
				k = nil
			}
			// close connection to router
			c.Close()
		}
	}
	if err == nil {
		// file should be there if it wasn't before
		_, err = os.Stat(fname)
		if err == nil {
			// file exists
			var f io.ReadCloser
			// open file
			f, err = os.Open(fname)
			if err == nil {
				// read keys
				r := bufio.NewReader(f)
				s.k = new(samKeys)
				// public component
				var addr string
				addr, err = r.ReadString(10)
				s.k.addr = I2PAddr(addr)
				if err == nil {
					// private component
					s.k.priv, err = r.ReadString(10)
				}
				// close key file
				f.Close()
				if err != nil {
					// clear keys because there was an error reading them
					s.k = nil
				}
			}
		}
	}
	// err != nil if file could not be opened or keys could not be generated
	return
}

// implements net.Listener
func (s *samSession) Accept() (c net.Conn, err error) {
	var nc net.Conn
	nc, err = s.connect("")
	if err == nil {
		_, err = fmt.Fprintf(nc, "STREAM ACCEPT ID=%s SILENT=false\n", s.name)
		if err == nil {
			r := bufio.NewReader(nc)
			var line string
			line, err = r.ReadString(10)
			if err == nil {
				if strings.HasPrefix(line, "STREAM STATUS RESULT=OK") {
					line, err = r.ReadString(10)
					if err == nil {
						c = &samConn{
							l: s.k.addr,
							r: I2PAddr(strings.Trim(line, "\n")),
							c: nc,
						}
						return
					}
				} else {
					err = errors.New("invalid line: " + line)
				}
			}
		}
	}
	// error
	nc.Close()
	return
}

// implements net.Listener and net.PacketConn
func (s *samSession) Addr() net.Addr {
	return s.k.addr
}

// implements net.Listener and net.PacketConn
func (s *samSession) Close() (err error) {
	err = s.c.Close()
	return
}

// implements i2p.Session
func (s *samSession) B32() string {
	return s.k.addr.DestHash().String()
}

// implements i2p.Session
func (s *samSession) Dial(network, addr string) (c net.Conn, err error) {
	c, err = s.connect("")
	if err == nil {
		// connected
		var a I2PAddr
		// strip out port for now
		if strings.Count(addr, ":") > 0 {
			addr = addr[:strings.Index(addr, ":")]
		}
		if strings.HasSuffix(addr, ".i2p") {
			// do lookup since it looks like an i2p address
			a, err = s.LookupI2P(addr)
			if err != nil {
				// lookup failed
				c.Close()
				c = nil
				return
			}
		} else if strings.Count(addr, ".") == 0 {
			// looks valid
			a = I2PAddr(addr)
		} else {
			// invalid address
			err = errors.New("invalid address: " + addr)
			c.Close()
			c = nil
			return
		}
		// send connect
		_, err = fmt.Fprintf(c, "STREAM CONNECT ID=%s DESTINATION=%s SILENT=false\n", s.name, a.String())
		if err == nil {
			// connect sent
			r := bufio.NewReader(c)
			var line string
			// read connect reply
			line, err = r.ReadString(10)
			if err == nil {
				// parse reply
				sc := bufio.NewScanner(strings.NewReader(line))
				sc.Split(bufio.ScanWords)
				for sc.Scan() {
					switch sc.Text() {
					case "STREAM":
						continue
					case "STATUS":
						continue
					case "RESULT=OK":
						// success
						c = &samConn{
							c: c,
							l: s.k.addr,
							r: a,
						}
						return c, nil
					default:
						// fail
						c.Close()
						err = errors.New(sc.Text())
						break
					}
				}
				return
			} else {
				// connect reply read failed
				c.Close()
				c = nil
			}
		} else {
			// send connect failed
			c.Close()
			c = nil
		}
	}
	return
}

// implements i2p.Session
func (s *samSession) Lookup(name string) (a net.Addr, err error) {
	a, err = s.LookupI2P(name)
	return
}

func (s *samSession) LookupI2P(name string) (a I2PAddr, err error) {
	s.cm.Lock()
	defer s.cm.Unlock()
	c := s.c
	if err == nil {
		_, err = fmt.Fprintf(c, "NAMING LOOKUP NAME=%s\n", name)
		if err == nil {
			r := bufio.NewReader(c)
			var line string
			line, err = r.ReadString(10)
			if err == nil {
				sc := bufio.NewScanner(strings.NewReader(line))
				sc.Split(bufio.ScanWords)
				for sc.Scan() {
					t := sc.Text()
					if t == "RESULT=OK" || t == "NAMING" || t == "REPLY" {
						continue
					} else if t == "NAME="+name || t == "NAME=ME" {
						continue
					} else if strings.HasPrefix(t, "VALUE=") {
						a = I2PAddr(t[6:])
						break
					} else {
						err = errors.New(line)
					}
				}
			}
		}
	}
	return
}

// create stream session
// ensurekeys must be called before
func (s *samSession) stream() (ss StreamSession, err error) {
	_, err = fmt.Fprintf(s.c, "SESSION CREATE STYLE=STREAM ID=%s DESTINATION=%s\n", s.name, s.k.priv)
	r := bufio.NewReader(s.c)
	var line string
	line, err = r.ReadString(10)
	if strings.HasPrefix(line, "SESSION STATUS RESULT=OK DESTINATION=") {
		// we good
		ss = s
	} else {
		err = errors.New(line)
	}
	return
}

// create a new uninitialized session
func newSession(name string) (s *samSession) {
	s = &samSession{
		name: name,
		minv: "3.0",
		maxv: "3.0",
	}
	return
}

func newSessionEasy(addr, keyfile string) (s *samSession, err error) {
	name := randStr(10)
	s = newSession(name)
	// dial out
	s.c, err = s.connect(addr)
	if err == nil {
		err = s.EnsureKeyfile(keyfile)
		if err == nil {
			// we gud
			return
		}
		s.Close()
	}
	s = nil
	return
}

// create a new stream session with i2p "the easy way"
func NewStreamSessionEasy(addr, keyfile string) (session StreamSession, err error) {
	var s *samSession
	s, err = newSessionEasy(addr, keyfile)
	if err == nil {
		// create stream session
		session, err = s.stream()
	}
	return
}

// create a new packet session with i2p "the easy way"
func NewPacketSessionEasy(addr, keyfile string) (session PacketSession, err error) {
	var s *samSession
	s, err = newSessionEasy(addr, keyfile)
	if err == nil {
		// parse addresses
		var host, port string
		host, port, err = net.SplitHostPort(addr)
		if err == nil {
			var p int
			p, err = strconv.Atoi(port)
			if err == nil {
				p--
				var ip *net.IPAddr
				ip, err = net.ResolveIPAddr("ip", host)
				if err == nil {
					var samaddr *net.UDPAddr
					samaddr, err = net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:%d", ip.IP.To16().String(), p))
					if err == nil {
						// find address that fits sam address
						var ifaddrs []net.Addr
						ifaddrs, err = net.InterfaceAddrs()
						if err == nil {
							for _, ifaddr := range ifaddrs {
								netip, ipnet, _ := net.ParseCIDR(ifaddr.String())
								if ipnet != nil && ipnet.Contains(ip.IP) {
									// found an address to bind to
									// create session
									session, err = s.packet(samaddr, netip)
									return
								}
							}
							err = errors.New("cannot find network address for " + host)
						}
					}
				}
			}
		}
	}
	return
}

// backwards compat
var NewSessionEasy = NewStreamSessionEasy

// backwards compat
//var NewSession = NewStreamSession
