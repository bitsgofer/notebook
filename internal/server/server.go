package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gorillahandlers "github.com/gorilla/handlers"
	"golang.org/x/crypto/acme/autocert"
	"k8s.io/klog/v2"
)

type Config struct {
	BlogRoot               string
	UseHTTPSOnly           bool
	LetsEncryptAdminEmail  string
	LetsEncryptDomains     []string
	InsecureHTTPListenAddr string // main listen addr (e.g: 80 or 443)
	MetricsPort            int    // port for Prometheus metrics + pprof
}

// Server wraps config, ACME autocert manager and httpServer for static files.
type Server struct {
	cfg             *Config
	autocertManager *autocert.Manager
	httpServer      *http.Server
}

func New(cfg *Config) (*Server, error) {
	if err := canonizeConfigAndValidate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config; err= %w", err)
	}

	manager := autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       autocert.DirCache(cfg.BlogRoot),
		HostPolicy:  autocert.HostWhitelist(cfg.LetsEncryptDomains...),
		Email:       cfg.LetsEncryptAdminEmail,
		RenewBefore: 30 * 24 * time.Hour,
	}
	httpServer := &http.Server{
		Addr:      cfg.InsecureHTTPListenAddr,
		TLSConfig: manager.TLSConfig(),
	}
	srv := &Server{
		cfg:             cfg,
		autocertManager: &manager,
		httpServer:      httpServer,
	}

	// The chain of HTTP handlers (i.e "HTTP middlewares") (e.g: metrics, logging, gzip, etc)
	// should be specified here.
	var handler http.Handler
	handler = srv.contentHandler()
	handler = gorillahandlers.CompressHandlerLevel(handler, gzip.BestCompression)
	handler = setCacheHeaderHandler(handler)
	srv.httpServer.Handler = handler

	return srv, nil
}

func (srv *Server) Run() {
	if srv.cfg.UseHTTPSOnly {
		srv.serveHTTPSWithACME()
		return
	}

	srv.serveHTTPInsecure()
}

// canonizeConfigAndValidate validate the given *Config.
// It will also update the config in a few places
// (e.g: use absolute path instead of relative path).
func canonizeConfigAndValidate(cfg *Config) error {
	blogRootAbsPath, err := filepath.Abs(cfg.BlogRoot)
	if err != nil {
		return fmt.Errorf("cannot get absolute path to .BlogRoot; err= %w", err)
	}
	cfg.BlogRoot = blogRootAbsPath
	klog.V(4).Infof(".BlogRoot= %q", cfg.BlogRoot)

	if _, err := os.Stat(cfg.BlogRoot); os.IsNotExist(err) {
		return fmt.Errorf(".BlogRoot does not exist")
	}

	if !cfg.UseHTTPSOnly {
		return nil
	}

	// additional checks if using HTTPS
	if len(cfg.LetsEncryptDomains) < 1 {
		return fmt.Errorf("need to have at least one domain for the Let's Encrypt to issue certificate (e.g: the blog domain")
	}
	if len(cfg.LetsEncryptAdminEmail) < 1 {
		return fmt.Errorf("need to specify admin email for Let's Encrypt")
	}

	return nil
}

// errHandlerFunc returns a handler func for a particular status code.
func (srv *Server) errHandlerFunc(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := srv.cfg.BlogRoot + "/error.html"
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			klog.Fatalf("cannot load error file %q; err= %q", filePath, err)
		}

		const placeholderCode = "HTTPStatusCode"
		const placeholderMsg = "ErrorMessageForHTTPStatusCode"
		realStatusCode := fmt.Sprintf("%d", statusCode)
		realErrorMsg := http.StatusText(statusCode)
		b = bytes.Replace(b, []byte(placeholderCode), []byte(realStatusCode), 1)
		b = bytes.Replace(b, []byte(placeholderMsg), []byte(realErrorMsg), 1)

		w.WriteHeader(statusCode)
		w.Write(b)
	}
}

// contentHandler returns the blog content handler, which should take care
// of all request that are not ACME's HTTP-01 challenge (https://letsencrypt.org/docs/challenge-types/#http-01-challenge).
func (srv *Server) contentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			srv.errHandlerFunc(http.StatusNotImplemented)(w, r)
			klog.Errorf("a %s request was received (not supported)", r.Method)
			return
		}

		// if we panic while serving, try to serve an 500 page
		defer func() {
			recoveredErr := recover()
			if recoveredErr != nil {
				srv.errHandlerFunc(http.StatusInternalServerError)(w, r)
				klog.Errorf("caught a panic; err= %q", recoveredErr)
			}
		}()

		// find file to serve from request path
		filePath := srv.cfg.BlogRoot + srv.findFileFromRequestPath(r)
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				srv.errHandlerFunc(http.StatusNotFound)(w, r)
				klog.Errorf("%q requested; err= not found", filePath)
				return
			}

			srv.errHandlerFunc(http.StatusInternalServerError)(w, r)
			klog.Errorf("os.Stat(%q) failed; err= %q", filePath, err)
			return
		}

		// serve it
		http.ServeFile(w, r, filePath)
		klog.V(2).Infof("served file %q", filePath)
	}
}

func setCacheHeaderHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		switch strings.ToLower(filepath.Ext(r.URL.Path)) {
		case ".css",
			".js",
			".ico",
			".jpg", ".jpeg", ".gif", ".png", ".svg":
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}

		h.ServeHTTP(w, r)
	})
}

// serveHTTPInsecure listens to HTTP request on .Config.HTTPListenAddr and serve
// static files.
func (srv *Server) serveHTTPInsecure() {
	if err := srv.httpServer.ListenAndServe(); err != nil {
		klog.Fatalf("HTTP server failed; err= %q", err)
	}
}

// serveHTTPSWithACME listens to HTTPS requests on :443 and serve static files.
// It will also listen for HTTP requests on :80, handling ACME's HTTP-01 challenge
// and redirect other requests to :443.
func (srv *Server) serveHTTPSWithACME() {
	// in another goroutine, listen for HTTP requests on :80
	go func() {
		httpHandler := srv.autocertManager.HTTPHandler(nil)
		if err := http.ListenAndServe(":http", httpHandler); err != nil {
			klog.Fatalf("HTTP server failed; err= %q", err)
		}
	}()

	// blog is served on :443
	srv.httpServer.Addr = ":https" // override the default (.cfg.InsecureHTTPListenAddr)
	if err := srv.httpServer.ListenAndServeTLS("", ""); err != nil {
		klog.Fatalf("HTTPS server failed; err= %q", err)
	}
}

// findFileFromRequestPath returns the static file to serve based on the request's path.
func (srv *Server) findFileFromRequestPath(r *http.Request) string {
	if r.URL.Path == "/" {
		return "/index.html"
	}

	switch strings.ToLower(filepath.Ext(r.URL.Path)) {
	// selected extensions => <path>.<extension>
	case ".html",
		".css",
		".js",
		".ico",
		".jpg", ".jpeg", ".gif", ".png", ".svg":
		return r.URL.Path
	default: // no extension => articles => <path>.html
	}

	return strings.TrimRight(r.URL.Path, "/") + ".html"
}
