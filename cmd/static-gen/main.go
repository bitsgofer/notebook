package main

import (
	"flag"
	"os"

	"github.com/exklamationmark/glog"

	"github.com/exklamationmark/notebook/internal/staticgen"
)

func main() {
	var postDir string
	var postTemplate string
	var outDir string

	flag.StringVar(&postDir, "dir", "_posts", "post directory")
	flag.StringVar(&postTemplate, "template", "template.html", "html template")
	flag.StringVar(&outDir, "out", "out", "output directory")

	flag.Parse()
	defer glog.Flush()

	if err := staticgen.Generate(postDir, postTemplate, outDir); err != nil {
		glog.Errorf("cannot generate static files in %q", postDir)
		os.Exit(1)
	}
}
