package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/exklamationmark/glog"

	"github.com/exklamationmark/notebook/internal/notebook"
)

func main() {
	var root string
	var addr string

	flag.StringVar(&addr, "addr", ":8080", "HTTP addr")
	flag.StringVar(&root, "root", "/var/www/notebook/", "root of notebook")

	flag.Parse()
	defer glog.Flush()

	nb, err := notebook.NewNotebook(root)
	if err != nil {
		glog.Errorf("cannot create notebook from %q, err= %v", root, err)
		os.Exit(1)
	}

	// http.HandleFunc("/nb1", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	nb.RenderHTML(w)
	// })
	http.HandleFunc("/", nb.HTTPHandler())

	glog.Warningf("listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		glog.Errorf("failed to run http server, err= %v", err)
		os.Exit(1)
	}
}
