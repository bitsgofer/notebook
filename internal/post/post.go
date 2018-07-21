package post

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"github.com/exklamationmark/glog"
	"github.com/pkg/errors"
	markdown "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

type Metadata struct {
	Author      string
	PublishedAt time.Time
	Title       string
	Slug        string
	Sticky      bool
	Tags        []string
}

type metdataYAML struct {
	Author      string   `yaml:"author"`
	PublishedAt string   `yaml:"published"`
	Title       string   `yaml:"title"`
	Slug        string   `yaml:"slug"`
	Sticky      bool     `yaml:"sticky"`
	Tags        []string `yaml:"tags"`
}

type Post struct {
	Filename string
	Metadata
	HTML template.HTML
}

func New(fname string) (*Post, error) {
	glog.V(1).Infof("parsing %q", fname)

	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %q", err)
	}

	rawMeta, rawMarkdown, err := splitRaw(raw)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("metadata= %s", string(rawMeta))
	glog.V(1).Infof("markdown= %s", string(rawMarkdown))

	meta, err := decodeMetadata(rawMeta)
	if err != nil {
		return nil, err
	}
	html := template.HTML(string(markdown.Run(rawMarkdown)))

	absFilename, err := filepath.Abs(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get absolute pat to %q", fname)
	}

	return &Post{
		Filename: absFilename,
		Metadata: meta,
		HTML:     html,
	}, nil
}

func (p *Post) CanonicalPath() string {
	yyyy, mm, dd := p.Metadata.PublishedAt.Date()
	return fmt.Sprintf("%04d/%02d/%02d/%s", yyyy, mm, dd, p.Metadata.Slug)
}

func (p *Post) ParentPath() string {
	yyyy, mm, dd := p.Metadata.PublishedAt.Date()
	return fmt.Sprintf("%04d/%02d/%02d", yyyy, mm, dd)
}

var (
	sep = []byte("---")
)

func splitRaw(raw []byte) (meta, markdown []byte, err error) {
	first := bytes.Index(raw, sep)
	if first == -1 {
		return nil, nil, errors.Errorf("missing metadata section")
	}
	second := bytes.Index(raw[first+3:], sep)
	if second == -1 {
		return nil, nil, errors.Errorf("missing metadata section")
	}
	second += first + len(sep) // changed to offset w.r.t raw

	meta = raw[first+len(sep) : second]
	markdown = raw[second+len(sep):]

	return meta, markdown, nil
}

func decodeMetadata(raw []byte) (Metadata, error) {
	dec := yaml.NewDecoder(bytes.NewBuffer(raw))
	var base metdataYAML
	if err := dec.Decode(&base); err != nil {
		return Metadata{}, errors.Wrapf(err, "cannot decode metadata")
	}

	publishedAt, err := time.Parse(time.RFC3339, base.PublishedAt)
	if err != nil {
		return Metadata{}, errors.Wrapf(err, "published time=%s is not in RFC3339 format", base.PublishedAt)
	}

	sort.Strings(base.Tags)

	return Metadata{
		Author:      base.Author,
		Title:       base.Title,
		Slug:        base.Slug,
		PublishedAt: publishedAt.UTC(),
		Sticky:      base.Sticky,
		Tags:        base.Tags, // moved
	}, nil
}
