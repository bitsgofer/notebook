package notebook

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/exklamationmark/notebook/internal/md2http"
	"github.com/pkg/errors"
)

type notebook struct {
	name     string
	articles []*article
}

type article struct {
	name     string
	editedAt time.Time
	html     []byte
}

func newArticle(path string, info os.FileInfo) (*article, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", path)
	}
	defer f.Close()

	var buf bytes.Buffer
	if err := md2http.Convert(f, &buf); err != nil {
		return nil, errors.Wrapf(err, "cannot convert content of %q to HTML", path)
	}

	return &article{
		name:     info.Name(),
		editedAt: info.ModTime(),
		html:     buf.Bytes(),
	}, nil
}

func walkNotebook(articles *[]*article) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "cannot walk article file: %q", path)
		}
		if info.IsDir() {
			return nil
		}

		atc, err := newArticle(path, info)
		if err != nil {
			return errors.Wrapf(err, "cannot create article")
		}

		(*articles) = append(*articles, atc)

		return nil
	}
}

func NewNotebook(path string) (*notebook, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "no stat for notebook root: %q", path)
	}

	var articles []*article
	if err := filepath.Walk(path, walkNotebook(&articles)); err != nil {
		return nil, errors.Wrapf(err, "cannot walk notebook's root: %q", path)
	}

	return &notebook{
		name:     info.Name(),
		articles: articles,
	}, nil
}

func (nb *notebook) RenderHTML(w io.Writer) {
	w.Write([]byte("<h1>"))
	w.Write([]byte(nb.name))
	w.Write([]byte("</h1>\n"))

	w.Write([]byte("<ul>"))
	for _, atc := range nb.articles {
		w.Write([]byte("<div><h2>"))
		w.Write([]byte(atc.name))
		w.Write([]byte("</h2>\n"))
		w.Write([]byte("<p>Edited at: "))
		w.Write([]byte(atc.editedAt.Format(time.RFC3339)))
		w.Write([]byte("</p>\n"))
		w.Write([]byte("</div>\n"))
	}
	w.Write([]byte("</ul>\n"))
}
