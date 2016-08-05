package model

import (
	"bytes"
	"crypto/rand"
	"bds/lib/base91"
	"bds/lib/cryptonight"
	"io"
	"strings"
)

// a hashed+salted login cred
type LoginCred string

const login_cred_delim = " "

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
	d := make([]byte, len(data)+len(salt))
	copy(d, data)
	copy(d[:len(data)], salt)
	r := cryptonight.HashBytes(d)
	h = r[:]
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
