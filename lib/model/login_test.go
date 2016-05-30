package model

import (
	"testing"
)

func TestLoginCredHash(t *testing.T) {
	secret := "password"
	cred := NewLoginCred(secret)
	t.Log(cred)
	t.Log(cred.hash())
	t.Log(cred.salt())
	if !cred.Check(secret) {
		t.Fail()
	}
	if cred.Check("not correct") {
		t.Fail()
	}
}
