package i2p

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
)

var (
	i2pB64enc *base64.Encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-~")
	i2pB32enc *base32.Encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")
)

// base64 long form of i2p destination
type I2PAddr string

// an i2p destination hash, the .b32.i2p address if you will
type I2PDestHash [32]byte

// get string representation of i2p dest hash
func (h I2PDestHash) String() string {
	b32addr := make([]byte, 56)
	i2pB32enc.Encode(b32addr, h[:])
	return string(b32addr[:52]) + ".b32.i2p"
}

// Returns the base64 representation of the I2PAddr
func (a I2PAddr) Base64() string {
	return string(a)
}

// Returns the I2P destination (base64-encoded)
func (a I2PAddr) String() string {
	return string(a)
}

// return base32 i2p desthash
func (a I2PAddr) DestHash() (dh I2PDestHash) {
	hash := sha256.New()
	b, _ := a.ToBytes()
	hash.Write(b)
	digest := hash.Sum(nil)
	copy(dh[:], digest)
	return
}

// return .b32.i2p address
func (a I2PAddr) Base32() string {
	return a.DestHash().String()
}

// decode to i2p address to raw bytes
func (a I2PAddr) ToBytes() (d []byte, err error) {
	buf := make([]byte, i2pB64enc.DecodedLen(len(a)))
	_, err = i2pB64enc.Decode(buf, []byte(a)) 
	if err == nil {
		d = buf
	}
	return
}

// Returns "I2P"
func (a I2PAddr) Network() string {
	return "I2P"
}

