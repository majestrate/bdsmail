package mailutil

import (
	"github.com/majestrate/bdsmail/lib/util"
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

func WriteBounceMail(wr io.Writer, recip, reason string, msg io.Reader) (err error) {
	b := util.RandStr(20)
	textPart := make(textproto.MIMEHeader)
	textPart.Set("Content-Type", "text/plain")
	bodyPart := make(textproto.MIMEHeader)
	bodyPart.Set("Content-Type", "application/octect-stream")
	c := textproto.NewWriter(bufio.NewWriter(wr))
	c.PrintfLine("Subject: failed to deliver mail")
	c.PrintfLine("From: postmaster@localhost")
	c.PrintfLine("To: %s", recip)
	c.PrintfLine("MIME-Version: 1.0")
	c.PrintfLine("Content-Type: multipart/mixed; boundary=%s", b)
	mw := multipart.NewWriter(wr)
	mw.SetBoundary(b)
	var w io.Writer
	w, err = mw.CreatePart(textPart)
	if err == nil {
		_, err = fmt.Fprintf(w, "your message failed to be delivered: %s", reason)
		if err == nil {
			w, err = mw.CreatePart(bodyPart)
			if err == nil {
				_, err = io.Copy(w, msg)
			}
		}
	}
	mw.Close()
	return
}
