package blogserver

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/exklamationmark/glog"
	"golang.org/x/crypto/acme/autocert"
)

type config struct {
	rootDir string
}

type server struct {
	config
	acmeManager *autocert.Manager
}

func New(rootDir, adminEmail string, domains ...string) (*server, error) {
	c := config{
		rootDir: rootDir,
	}

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(c.rootDir),
		HostPolicy: autocert.HostWhitelist(domains...),
		Email:      adminEmail,
	}

	return &server{
		config:      c,
		acmeManager: &manager,
	}, nil
}

func (srv *server) httpHandler() http.Handler {
	return srv.acmeManager.HTTPHandler(nil)
}

func (srv *server) httpsHandler() http.Handler {
	return http.HandlerFunc(blogHandler(srv.config.rootDir))
}

func (srv *server) HTTPServer() *http.Server {
	return &http.Server{
		Handler: srv.httpHandler(),
	}
}

func (srv *server) HTTPSServer() *http.Server {
	return &http.Server{
		Handler: srv.httpsHandler(),
		TLSConfig: &tls.Config{
			GetCertificate: srv.acmeManager.GetCertificate,
		},
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

func fileToServe(rootDir, path string) string {
	ext := filepath.Ext(path)
	switch {
	case path == "/":
		return rootDir + "/index.html"
	case ext == ".ico" || ext == ".css" || ext == ".js" || ext == ".html":
		return rootDir + path
	default:
		return rootDir + strings.TrimRight(path, "/") + ".html"
	}
}

func blogHandler(rootDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			serveErrPage(w, r, http.StatusBadRequest)
			return
		}

		fname := fileToServe(rootDir, r.URL.Path)
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
