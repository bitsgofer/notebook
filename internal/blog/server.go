package blog

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/exklamationmark/glog"
	"github.com/pkg/errors"
	"golang.org/x/crypto/acme/autocert"

	"github.com/exklamationmark/notebook/internal/middlewares/redirect"
)

type config struct {
	htmlDir      string
	redirections redirect.Redirections
}

type opt func(*config)

type Server struct {
	config
	acmeManager *autocert.Manager
	handler     http.Handler
}

func New(htmlDir, adminEmail string, domains []string, opts ...opt) (*Server, error) {
	absHTMLDir, err := filepath.Abs(htmlDir)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot find absolute path to %q", htmlDir)
	}

	c := config{htmlDir: absHTMLDir}
	for _, opt := range opts {
		opt(&c)
	}

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(c.htmlDir),
		HostPolicy: autocert.HostWhitelist(domains...),
		Email:      adminEmail,
	}

	next := http.HandlerFunc(blogHandler(c.htmlDir))
	handler, err := redirect.NewHandler(next, c.redirections, domains...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create handler")
	}

	return &Server{
		config:      c,
		acmeManager: &manager,
		handler:     handler,
	}, nil
}

func Redirect(rds redirect.Redirections) func(*config) {
	return func(c *config) {
		c.redirections = rds
	}
}

func (srv *Server) HTTPRedirectHandler() http.Handler {
	return srv.acmeManager.HTTPHandler(nil)
}

func (srv *Server) BlogHandler() http.Handler {
	return srv.handler
}

func (srv *Server) TLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: srv.acmeManager.GetCertificate,
	}
}

var (
	defaultResponses = map[int][]byte{
		http.StatusInternalServerError: []byte("<html><h1>Internal server error</h1><p>Sorry, something went wrong</p></html>"),
		http.StatusNotFound:            []byte("<html><h1>Not found</h1><p>Sorry, but our princess is in another castle</p></html>"),
		http.StatusBadRequest:          []byte("<html><h1>Bad request</h1><p>Sorry, this we can't serve this</p></html>"),
	}
)

func serveErrPage(w http.ResponseWriter, r *http.Request, status int) {
	b, exist := defaultResponses[status]
	if !exist {
		glog.Errorf("tried to render non-default error page, status= %d", status)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(defaultResponses[http.StatusInternalServerError])
		return
	}

	w.WriteHeader(status)
	w.Write(b)
}

var (
	servableExts = []string{
		".html",
		".css",
		".js",
		".ico",
		".jpg",
		".jpeg",
		".png",
		".gif",
		".pdf",
		".svg",
		".woff",
		".woff2",
		".ttf",
	}
)

func isServable(ext string) bool {
	for _, servable := range servableExts {
		if servable == ext {
			return true
		}
	}

	return false
}

func fileToServe(htmlDir, path string) string {
	ext := filepath.Ext(path)
	switch {
	case path == "/":
		return htmlDir + "/index.html"
	case isServable(ext):
		return htmlDir + path
	default:
		return htmlDir + strings.TrimRight(path, "/") + ".html"
	}
}

func blogHandler(htmlDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			serveErrPage(w, r, http.StatusBadRequest)
			return
		}

		fname := fileToServe(htmlDir, r.URL.Path)
		if _, err := os.Stat(fname); err != nil {
			if os.IsNotExist(err) {
				glog.Errorf("%q requested but not found", fname)
				serveErrPage(w, r, http.StatusNotFound)
				return
			}

			glog.Errorf("os.Stat(%q) failed, err= %v", fname, err)
			serveErrPage(w, r, http.StatusInternalServerError)
			return
		}

		serveFile(w, r, fname)
		glog.V(0).Infof("served %q", fname)
	}
}

var serveFile = func(w http.ResponseWriter, r *http.Request, fname string) {
	http.ServeFile(w, r, fname)
}
