package staticgen

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
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

	posts := make([]*post.Post, 0, 20)
	if err := filepath.Walk(postDir, processPost(htmlDir, tmpl, &posts)); err != nil {
		return errors.Wrapf(err, "cannot process all posts")
	}

	generateIndex(htmlDir, tmpl, posts)

	return nil
}

func generateIndex(outDir string, tmpl *template.Template, posts []*post.Post) error {
	sort.Slice(posts, func(i, j int) bool {
		first := posts[i].Metadata.PublishedAt
		second := posts[j].Metadata.PublishedAt
		return first.Before(second)
	})

	var buf bytes.Buffer
	buf.WriteString("<h2>Index</h2>\n")
	buf.WriteString("<ul>\n")
	for _, p := range posts {
		path := "/" + p.CanonicalPath()
		buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", path, p.Metadata.Title))
	}
	buf.WriteString("</ul>\n")
	p := &post.Post{
		Metadata: post.Metadata{
			Title: "Bitsgofer",
		},
		HTML: template.HTML(buf.String()),
	}

	fname := outDir + "/index.html"
	if err := renderPost(fname, p, tmpl); err != nil {
		return errors.Wrapf(err, "cannot render index")
	}

	glog.V(0).Infof("generated %s", fname)
	return nil
}

func processPost(outDir string, tmpl *template.Template, posts *[]*post.Post) func(string, os.FileInfo, error) error {
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

		*posts = append(*posts, post)

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
