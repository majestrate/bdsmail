package i2p

import (
  "net"
)

// a session with an i2p router
type Session interface {
	// get printable b32 address of this session
	B32() string
	// ensure a keyfile exists
	EnsureKeyfile(fpath string) (err error)
	// lookup i2p name
	LookupI2P(name string) (I2PAddr, error)
}

//
// session for accepting and creating reliable stream oriented connections over i2p from 1 anonymous endpoint
//
type StreamSession interface {
	// implements i2p.Session
	Session
	// implements net.Listener
	net.Listener
  // Dial out to i2p from this session
  // same api as net.Dialer.Dial
	// does not dial out to anything except i2p addresses
	// will fail if attempting to connect to something not accessable via the i2p network layer
  Dial(network, addr string) (net.Conn, error)
}

//
// session for sending and receiving large semi-reliable messages over i2p from 1 anonymous endpoint
//
type PacketSession interface {
	// implements i2p.Session
	Session
	// implements net.PacketConn
	net.PacketConn
}
