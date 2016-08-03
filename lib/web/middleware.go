package web

import (
	"github.com/gorilla/mux"
	"github.com/majestrate/bdsmail/lib/db"
	"github.com/majestrate/bdsmail/lib/web/admin"
	"github.com/majestrate/bdsmail/lib/web/webmail"
	"net/http"
)

// create middleware for web ui
func NewMiddleware(assetsdir string, dao db.DB) http.Handler {
	r := mux.NewRouter()
	// admin actions
	r.Handle("/admin", admin.New(dao))
	// mail actions
	r.Handle("/mail", webmail.New(dao))
	// file server
	r.Handle("/", http.FileServer(http.Dir(assetsdir)))
	return r
}
