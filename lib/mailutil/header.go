package mailutil

import (
	"fmt"
	"io"
	"time"
)

func WriteRecvHeader(w io.Writer, to, remoteAddr, remoteName, hostname, appname string) (err error) {
	now := time.Now().Format("Mon, _2 Jan 2006 22:04:05 -0000 (UTC)")
	_, err = fmt.Fprintf(w, "Received: from %s (%s [127.0.0.1])\r\n", remoteName, remoteAddr)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(w, "        by %s (%s) with SMTP\r\n", hostname, appname)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(w, "        for <%s>; %s\r\n", to, now)
	return
}
