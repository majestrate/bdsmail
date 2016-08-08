package web

import (
	"bds/lib/db"
	"bds/lib/web/admin"
	"bds/lib/web/webmail"
	"net/http"
)

// create middleware for web ui
func NewMiddleware(assetsdir string, dao db.DB) http.Handler {
	r := newRouter()
	// admin actions
	r.Handle("/admin", admin.New(dao))
	// mail actions
	r.Handle("/mail", webmail.New(dao))
	// file server
	r.HandleDefault(http.FileServer(http.Dir(assetsdir)))
	return r
}
