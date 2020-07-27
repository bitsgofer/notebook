package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type Blog struct {
	cfg             Config
	autocertManager *autocert.Manager
	server          *http.Server
}

func New(cfg Config) *Blog {
	// TODO(mark): validate cfg

	manager := autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       autocert.DirCache(cfg.BlogRoot),
		HostPolicy:  autocert.HostWhitelist(cfg.LetsEncryptDomains...),
		Email:       cfg.LetsEncryptAdminEmail,
		RenewBefore: 30 * 24 * time.Hour,
	}

	// create http server for blog content + assets
	mux := http.NewServeMux()
	// mux.HandleFunc("/", blogHandler)
	server := http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	return &Blog{
		cfg:             cfg,
		autocertManager: &manager,
		server:          &server,
	}
}

func (s *Blog) Run() {
	blogHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			serveErrPage(w, r, http.StatusBadRequest)
			return
		}

		fname := fileToServe(s.cfg.BlogRoot, r.URL.Path)
		if _, err := os.Stat(fname); err != nil {
			if os.IsNotExist(err) {
				klog.Errorf("%q requested but not found", fname)
				serveErrPage(w, r, http.StatusNotFound)
				return
			}

			klog.Errorf("os.Stat(%q) failed, err= %v", fname, err)
			serveErrPage(w, r, http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, fname)
		klog.V(2).Infof("served %q", fname)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", blogHandler)

	blogSrv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := blogSrv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("blog server failed; err= %w", err))
		os.Exit(1)
	}
}

func serveErrPage(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	w.Write([]byte("Error, try again"))
}

func fileToServe(htmlDir, path string) string {
	if path == "/" {
		return htmlDir + "/index.html"
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".css", ".js", ".ico", ".jpg", ".jpeg", ".gif", ".png", ".svg":
		return htmlDir + path
	default:
		return htmlDir + strings.TrimRight(path, "/") + ".html"
	}
}
