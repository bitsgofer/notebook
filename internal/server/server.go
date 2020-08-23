package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type errHandler func(statusCode int) http.Handler

type Server struct {
	cfg             *Config
	autocertManager *autocert.Manager
	httpServer      *http.Server

	// serveErr errHandler
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

	srv := &Server{
		cfg:             cfg,
		autocertManager: &manager,
	}

	httpServer := &http.Server{
		Addr:      cfg.InsecureHTTPListenAddr,
		Handler:   srv.contentHandler(),
		TLSConfig: manager.TLSConfig(),
	}
	srv.httpServer = httpServer

	return srv, nil
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

		const placeholderCode = "PlaceholderStatusCode"
		const placeholderMsg = "PlaceholderErrorMessage"
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
//
// The chain of HTTP handlers (i.e "HTTP middlewares") (e.g: metrics, logging, gzip, etc)
// should be specified here, too.
func (srv *Server) contentHandler() *http.ServeMux {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			srv.errHandlerFunc(http.StatusNotImplemented)(w, r)
			klog.Errorf("a %s request was received (not supported)", r.Method)
			return
		}

		filePath := srv.cfg.BlogRoot + srv.findFileFromRequestPath(r)
		b, err := ioutil.ReadFile(filePath)
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

		w.WriteHeader(http.StatusOK)
		w.Write(b)
		klog.V(2).Infof("served file %q", filePath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return mux
}

// serveHTTPInsecure listens to HTTP request on .Config.HTTPListenAddr and serve
// static files.
func (srv *Server) serveHTTPInsecure() {
	if err := srv.httpServer.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("blog server failed; err= %w", err))
		os.Exit(1)
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
	if err := srv.httpServer.ListenAndServeTLS(":https", ""); err != nil {
		klog.Fatalf("HTTPS server failed; err= %q", err)
	}
}

func (srv *Server) Run() {
	if srv.cfg.UseHTTPSOnly {
		srv.serveHTTPSWithACME()
		return
	}

	srv.serveHTTPInsecure()
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
