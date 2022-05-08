package starttls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/textproto"
	"os"
	"time"
)

var ErrTlsNotSupported = errors.New("STARTTLS not supported")

// handle STARTTLS on connection
func HandleStartTLS(conn net.Conn, config *tls.Config) (econn *textproto.Conn, state tls.ConnectionState, err error) {
	if config == nil {
		err = ErrTlsNotSupported
	} else {
		// begin tls crap here
		tconn := tls.Server(conn, config)
		err = tconn.Handshake()
		state = tconn.ConnectionState()
		if err == nil {
			econn = textproto.NewConn(tconn)
			return
		} else {
			tconn.Close()
		}
	}
	return
}

// create base tls certificate
func NewTLSCert(org string, ca bool) x509.Certificate {
	return x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Date(9005, 1, 1, 1, 1, 1, 1, time.UTC),
		BasicConstraintsValid: true,
		IsCA:                  ca,
	}
}

// generate tls config, private key and certificate
func GenTLS(hostname, org, certfile, privkeyfile string, bits int) (tcfg *tls.Config, err error) {
	// check for private key
	if _, err = os.Stat(privkeyfile); os.IsNotExist(err) {
		err = nil
		// no private key, let's generate it
		k := NewTLSCert(org, false)
		var priv *rsa.PrivateKey
		priv, err = rsa.GenerateKey(rand.Reader, bits)
		if err == nil {
			serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 256)
			k.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
			k.DNSNames = append(k.DNSNames, hostname)
			k.Subject.CommonName = hostname
			if err == nil {
				var derBytes []byte
				derBytes, err = x509.CreateCertificate(rand.Reader, &k, &k, &priv.PublicKey, priv)
				var f io.WriteCloser
				f, err = os.Create(certfile)
				if err == nil {
					err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
					f.Close()
					if err == nil {
						f, err = os.Create(privkeyfile)
						if err == nil {
							err = pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
							f.Close()
						}
					}
				}
			}
		}
	}
	if err == nil {
		// we should have the key generated and stored by now
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(certfile, privkeyfile)
		if err == nil {
			tcfg = &tls.Config{
				CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
				Certificates: []tls.Certificate{cert},
			}
		}
	}
	return
}
