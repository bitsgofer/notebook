package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
	"k8s.io/klog/v2"
)

type Config struct {
	BlogRoot              string
	UseHTTPSOnly          bool
	LetsEncryptAdminEmail string
	LetsEncryptDomains    []string
	ListenAddr            string // main listen addr (e.g: 80 or 443)
	MetricsPort           int    // port for Prometheus metrics + pprof
}

type errHandler func(statusCode int) http.Handler

type Blog struct {
	cfg             Config
	autocertManager *autocert.Manager
	server          *http.Server

	// serveErr errHandler
}

func New(cfg Config) *Blog {
	// TODO(mark): validate cfg

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(cfg.BlogRoot),
		HostPolicy: autocert.HostWhitelist(cfg.LetsEncryptDomains...),
		// Email:       cfg.LetsEncryptAdminEmail,
		// RenewBefore: 30 * 24 * time.Hour,
	}

	blogRoot, err := filepath.Abs(cfg.BlogRoot)
	if err != nil {
		panic(fmt.Errorf("cannot find absolute path to %q; err= %w", cfg.BlogRoot, err))
	}
	klog.V(4).Infof("blogRoot= %q\n", blogRoot)
	errPage, _ := ioutil.ReadFile(blogRoot + "/error.html")
	serveErr := func(statusCode int) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Write(errPage)
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		toServe := blogRoot + fileToServe(r)

		if _, err := os.Stat(toServe); err != nil {
			if os.IsNotExist(err) {
				klog.Errorf("%q requested; err= not found", toServe)
				serveErr(http.StatusNotFound)(w, r)
				return
			}

			klog.Errorf("os.Stat(%q) failed; err= %q", toServe, err)
			serveErr(http.StatusInternalServerError)(w, r)
			return
		}

		http.ServeFile(w, r, toServe)
		klog.V(2).Infof("served file %q", toServe)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := http.Server{
		Addr:      cfg.ListenAddr,
		Handler:   mux,
		TLSConfig: manager.TLSConfig(),
	}

	return &Blog{
		cfg:             cfg,
		autocertManager: &manager,
		server:          &server,
	}
}

func (b *Blog) Run() {
	if b.cfg.UseHTTPSOnly {
		if err := b.server.ListenAndServeTLS("", ""); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("blog server failed; err= %w", err))
			os.Exit(1)
		}
		return
	}

	if err := b.server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("blog server failed; err= %w", err))
		os.Exit(1)
	}
}

// fileToServe returns the static file to serve based on the request.
func fileToServe(r *http.Request) string {
	if r.Method != http.MethodGet {
		return "error.html"
	}

	if r.URL.Path == "/" {
		return "/index.html"
	}

	switch strings.ToLower(filepath.Ext(r.URL.Path)) {
	// selected extensions => matching static file
	case ".html", ".css", ".js", ".ico", ".jpg", ".jpeg", ".gif", ".png", ".svg":
		return r.URL.Path
	default: // below
	}

	// no extension => articles => find the .html file
	return strings.TrimRight(r.URL.Path, "/") + ".html"
}
