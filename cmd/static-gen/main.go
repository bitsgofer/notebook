package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/exklamationmark/glog"
	"github.com/exklamationmark/notebook/internal/post"

	"github.com/pkg/errors"
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

	tmpl, err := template.ParseFiles(postTemplate)
	if err != nil {
		glog.Errorf("cannot parse template %q, err= %v", postTemplate, err)
		os.Exit(1)
	}

	outDir = strings.TrimRight(outDir, "/")

	// generate into temp
	if err := filepath.Walk(postDir, func(fname string, stat os.FileInfo, err error) error {
		if stat.IsDir() {
			return nil
		}

		baseName := stat.Name()
		ext := filepath.Ext(baseName)
		if !(ext == ".md" || ext == ".markdown") {
			return nil
		}
		glog.V(0).Infof("processing: %s", fname)

		post, err := post.New(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot create post from %q", fname)
		}

		parentPath := fmt.Sprintf("%s/%s", outDir, post.ParentPath())
		if err := os.MkdirAll(parentPath, 0776); err != nil {
			return errors.Wrapf(err, "cannot create parent directory %q", parentPath)
		}

		generatedFile := fmt.Sprintf("%s/%s.html", outDir, post.CanonicalPath())
		if err := renderPost(generatedFile, post, tmpl); err != nil {
			return errors.Wrapf(err, "cannot render post from %q", fname)
		}

		glog.V(0).Infof("generated %s", generatedFile)

		return nil
	}); err != nil {
		glog.Errorf("cannot process all files in %q, err= %v", postDir, err)
		os.Exit(1)
	}
}

func renderPost(fname string, p *post.Post, tmpl *template.Template) error {
	f, err := os.Create(fname)
	if err != nil {
		return errors.Wrapf(err, "cannot open file to write")
	}
	defer f.Close()

	if err := tmpl.Execute(f, p); err != nil {
		return errors.Wrapf(err, "cannot render post")
	}

	return nil
}
