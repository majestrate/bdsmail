package model

import (
	"github.com/majestrate/bdsmail/lib/base91"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"strings"
)

// a hashed+salted login cred
type LoginCred string

const login_cred_delim = " "

// how many iterations of sha256 to use
const num_sha_digest_iteration = 1024 * 1024 

// generate a new login cred, generates random salt and stores as hashed
func NewLoginCred(secret string) LoginCred {
	slt := make([]byte, 64)
	io.ReadFull(rand.Reader, slt)
	h := credHash([]byte(secret), slt)
	hs := string(base91.Encode(h))
	ss := string(base91.Encode(slt))
	return LoginCred(hs + login_cred_delim + ss)
}

// hash login credential with a salt
func credHash(data, salt []byte) (h []byte) {
	l := len(data) + len(salt)
	d := make([]byte, l + sha256.Size)
	copy(d, data)
	copy(d[:len(data)], salt)
	i := num_sha_digest_iteration
	for i > 0 {
		_h := sha256.Sum256(d)
		copy(d[l:], _h[:])
		i--
	}
	h = make([]byte, sha256.Size)
	copy(h, d[l:])
	return
}

// check if this password matches this login cred
func (cred LoginCred) Check(passwd string) (is bool) {
	h := cred.hash()
	s := cred.salt()
	r := credHash([]byte(passwd), s)
	if len(h) > 0 {
		is = bytes.Equal(h, r[:])
	}
	return
}

// get hash part
func (cred LoginCred) hash() (h []byte) {
	p := strings.Split(string(cred), login_cred_delim)
	if len(p) == 2 {
		h, _ = base91.Decode([]byte(p[0]))
	}
	return
}

// get salt part
func (cred LoginCred) salt() (s []byte) {
	p := strings.Split(string(cred), login_cred_delim)
	if len(p) == 2 {
		s, _ = base91.Decode([]byte(p[1]))
	}
	return
}
