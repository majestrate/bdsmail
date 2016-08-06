package sendmail

import (
	"net"
)

// dialer function
type Dialer func(string, string) (net.Conn, error)

// domain name resolver function
type Resolver func(string) (net.Addr, error)
