package cryptonight

import (
	"bytes"
	"testing"
)

func TestHashBytes(t *testing.T) {
	var b [1024]byte
	r1 := HashBytes(b[:])
	r2 := HashBytes(b[:])
	if ! bytes.Equal(r1[:], r2[:]) {
		t.Fail()
	}
}
