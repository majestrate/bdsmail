package web

import (
	"net/http"
	"strings"
)

// custom http router
type httpMux struct {
	routes         map[string]http.Handler
	defaultHandler http.Handler
}

func (m *httpMux) Handle(path string, handler http.Handler) {
	m.routes[path] = handler
}

func (m *httpMux) HandleDefault(handler http.Handler) {
	m.defaultHandler = handler
}

func (m *httpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := m.defaultHandler
	path := r.URL.String()
	prefix := strings.Split(path, "/")
	if len(prefix) > 1 {
		r, ok := m.routes[prefix[1]]
		if ok {
			route = r
		}
	}
	route.ServeHTTP(w, r)
}

func newRouter() *httpMux {
	return &httpMux{
		routes:         make(map[string]http.Handler),
		defaultHandler: http.HandlerFunc(http.NotFound),
	}
}
