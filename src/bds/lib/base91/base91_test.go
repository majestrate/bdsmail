package base91

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestBase91(t *testing.T) {
	n := 0
	for n < 1028 {
		buff := make([]byte, n)
		io.ReadFull(rand.Reader, buff[:])
		out := Encode(buff[:])
		t.Logf("%q -> %q", buff[:], out)
		buff2, err := Decode(out)
		if err == nil {
			if !bytes.Equal(buff2, buff[:]) {
				t.Logf("%q vs %q", buff2, buff)
				t.Fail()
			}
		} else {
			t.Logf("error: %s", err.Error())
			t.Fail()
		}
		n++
		buff = nil
	}
}
