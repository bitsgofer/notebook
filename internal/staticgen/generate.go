package staticgen

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/exklamationmark/glog"
	"github.com/exklamationmark/notebook/internal/post"
	"github.com/pkg/errors"
)

func Generate(postDir, postTemplate, htmlDir string) error {
	tmpl, err := template.ParseFiles(postTemplate)
	if err != nil {
		return errors.Wrapf(err, "cannot parse post template %q", postTemplate)
	}

	htmlDir = strings.TrimRight(htmlDir, "/")

	if err := filepath.Walk(postDir, processPost(htmlDir, tmpl)); err != nil {
		return errors.Wrapf(err, "cannot process all posts")
	}

	return nil
}

func processPost(outDir string, tmpl *template.Template) func(string, os.FileInfo, error) error {
	return func(fname string, stat os.FileInfo, err error) error {
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
			return err
		}

		glog.V(0).Infof("generated %s", generatedFile)
		return nil
	}
}

func renderPost(fname string, p *post.Post, tmpl *template.Template) error {
	f, err := os.Create(fname)
	if err != nil {
		return errors.Wrapf(err, "cannot open file to write")
	}
	defer f.Close()

	if err := tmpl.Execute(f, p); err != nil {
		return errors.Wrapf(err, "cannot execute template post")
	}

	return nil
}
