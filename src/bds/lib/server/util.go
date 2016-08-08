package server

import (
	"fmt"
	"regexp"
	"strings"
)

// given an email address get the i2p destination name for it
func parseFromI2PAddr(email string) (name string) {
	idx_at := strings.Index(email, "@")
	if strings.HasSuffix(email, ".b32.i2p") {
		name = email[idx_at+1:]
	} else if strings.HasSuffix(email, ".i2p") {
		idx_i2p := strings.LastIndex(email, ".i2p")
		name = fmt.Sprintf("smtp.%s.i2p", email[idx_at+1:idx_i2p])
	}
	name = strings.Trim(name, ",= \t\r\n\f\b")
	return
}

var re_email = regexp.MustCompile(`<?([a-zA-Z0-9]*@[a-zA-z0-9]*\.i2p)>?`)


func normalizeEmail(email string) (e string) {
	e = re_email.FindString(email)
	return
}
