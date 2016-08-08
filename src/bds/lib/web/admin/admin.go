package admin

import (
	"bds/lib/db"
	"net/http"
)

type Admin struct {
	d db.DB
}

// handle admin request
func (a *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func New(dao db.DB) *Admin {
	return &Admin{
		d: dao,
	}
}
