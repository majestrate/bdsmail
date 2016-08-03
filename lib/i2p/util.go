package i2p

import (
  "crypto/rand"
  "encoding/base64"
	"strings"
)

// generate a random string of n chars long
func randStr(n int) (str string) {
  buff := make([]byte, n)
  rand.Reader.Read(buff)
  s := base64.StdEncoding.EncodeToString(buff)
  b := make([]byte, len(s))
  copy(b, []byte(s))
  str = string(b)
	str = strings.Trim(str, "=")
  return
}
