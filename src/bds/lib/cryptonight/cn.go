package cryptonight

// #cgo CFLAGS: -maes
// #include "cn.h"
import "C"
import "unsafe"

type Result [32]byte

// do cryptonight hash function on a byte slice
func HashBytes(d []byte) (r Result) {
	l := len(d)
	b := make([]C.char, l)
	for i, c := range d {
		b[i] = C.char(c)
	}
	var cr [32]C.char
	bptr := unsafe.Pointer(&b[0])
	C.cn_slow_hash(bptr, C.size_t(l), &cr[0])
	for i, c := range cr {
		r[i] = byte(c)
	}
	return
}
