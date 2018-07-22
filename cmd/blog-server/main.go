package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/exklamationmark/glog"
)

func main() {
	var root string
	var addr string
	var tlsKeyFile string
	var tlsCertFile string

	flag.StringVar(&root, "root", "public_html", "root of blog")
	flag.StringVar(&addr, "addr", ":80", "listening addr")
	flag.StringVar(&tlsKeyFile, "tls.key", "/path/to/key", "SSL's private key file")
	flag.StringVar(&tlsCertFile, "tls.cert", "/path/to/cert", "SSL's private cert file")

	flag.Parse()

	srv := &Server{
		root: strings.TrimRight(root, "/"),
	}

	http.HandleFunc("/", srv.handler)
	glog.Infof(addr)
	if err := http.ListenAndServeTLS(addr, tlsCertFile, tlsKeyFile, nil); err != nil {
		glog.Errorf("cannot serve; err= %v", err)
	}
}

type Server struct {
	root string
}

func (srv *Server) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := strings.TrimLeft(r.URL.Path, "/")
	ext := filepath.Ext(path)
	if path == "" {
		file := fmt.Sprintf("%s/index.html", srv.root)
		serveFile(w, r, file)
		return
	}
	if ext == ".ico" || ext == ".css" || ext == ".js" || ext == ".html" {
		file := fmt.Sprintf("%s/%s", srv.root, path)
		serveFile(w, r, file)
		return
	}

	file := fmt.Sprintf("%s/%s.html", srv.root, strings.TrimRight(path, "/"))
	serveFile(w, r, file)

}

func serveFile(w http.ResponseWriter, r *http.Request, file string) {
	stat, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Errorf("file %q does not exist:", file)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<p>Our princess is in another castle</p>"))
			return
		}

		glog.Errorf("cannot get stat for %q; err= %v", file, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		glog.Errorf("cannot serve directory %q", file)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	glog.Infof("serving %s", file)
	http.ServeFile(w, r, file)
}
