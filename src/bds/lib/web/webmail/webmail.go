package webmail

import (
	"github.com/majestrate/bdsmail/lib/db"
	"net/http"
)

type WebMail struct {
	d db.DB
}

func (m *WebMail) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func New(dao db.DB) *WebMail {
	return &WebMail{
		d: dao,
	}
}
