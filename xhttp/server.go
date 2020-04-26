package xhttp

import (
	"net"
	"net/http"
	"path"
	"strings"
)

type UnsafeDynamicHandler struct {
	entries []func(host string, path string) http.Handler
}

func (m *UnsafeDynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := stripHostPort(r.Host)
	path := cleanPath(r.URL.Path)
	// scanf scans space-separated values
	path = strings.ReplaceAll(path, "/", " ")
	e := m.entries
	e = append(e, func(host string, path string) http.Handler {
		return http.NotFoundHandler()
	})
	for i := range e {
		h := e[i](host, path)
		if h == nil {
			continue
		}
		h.ServeHTTP(w, r)
		return
	}
}

func (m *UnsafeDynamicHandler) Append(f func(host string, path string) http.Handler) {
	m.entries = append(m.entries, f)
}

// stripHostPort returns h without any trailing ":<port>".
func stripHostPort(h string) string {
	// If no port on host, return unchanged
	if strings.IndexByte(h, ':') == -1 {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h // on error, return unchanged
	}
	return host
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
