package blogcontent

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

var (
	dns1123Regexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// ParseArticle parses a file into an Article.
func ParseArticle(r io.Reader) (*Article, error) {
	metadata, content, err := readFile(r)
	if err != nil {
		return nil, err
	}

	article := Article{
		content: content,
	}

	// unmarshal metadata
	buf := bytes.NewBuffer(metadata)
	dec := yaml.NewDecoder(buf)
	dec.KnownFields(true)
	if err := dec.Decode(&article.Metadata); err != nil {
		return nil, fmt.Errorf("cannot unmarshal metadata (yaml); err= %w", err)
	}

	// set .ID, .URL and .FileName
	hash := md5.Sum([]byte(article.Metadata.Title))
	article.ID = fmt.Sprintf("%x", hash[:])
	name := strings.ToLower(strings.ReplaceAll(article.Metadata.Title, " ", "-"))
	if !dns1123Regexp.MatchString(name) {
		return nil, fmt.Errorf("generated article name %q is not a DNS-safe string", name)
	}
	article.URL = fmt.Sprintf("/%s", name)
	article.FileName = fmt.Sprintf("%s.html", name)

	// render HTML .Content
	html5Content, err := ToHTML(article.content)
	if err != nil {
		return nil, fmt.Errorf("cannot render .Content as HTML5; err= %w", err)
	}
	article.Content = template.HTML(html5Content)
	// render HTML .Metadata.Summary
	html5Content, err = ToHTML([]byte(article.Metadata.RawSummary))
	if err != nil {
		return nil, fmt.Errorf("cannot render .Metadata.RawSummary as HTML5; err= %w", err)
	}
	article.Metadata.Summary = template.HTML(html5Content)

	return &article, nil
}

// readFile parses a file into byte slices of metadata and content.
func readFile(r io.Reader) (metadata, content []byte, err error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read all bytes; err= %w", err)
	}

	metadataEnd, contentStart, err := splitBySeparator(b)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot split post into (metadata, content); err= %w", err)
	}

	return b[:metadataEnd], b[contentStart:], nil
}

// splitBySeparator finds the locations of metadata and content in a blog post.
// Assuming each valid post have both metadata and content, it searchs for
// the separator `----` (rune '-' x4).
// If found, it returns the index where metadata ends + where content start.
// Otherwise, it returns an error.
func splitBySeparator(b []byte) (metadataEndPos, contentStartPos int, err error) {
	pos := -1
	n := len(b)
	klog.V(4).Infof("search for `----` in str= %q", b)

	for {
		pos++
		if pos >= n-1 {
			klog.V(4).Infof("pos= %d; n= %d => reached EOF", pos, n)
			return -1, -1, fmt.Errorf("separator not found")
			break
		}

		if b[pos] != '-' {
			continue
		}
		if pos+3 <= n-1 { // can peek 3 more bytes
			if b[pos+1] == '-' && b[pos+2] == '-' && b[pos+3] == '-' { // found
				if pos+4 > n-1 { // no content behind separator
					klog.V(4).Infof("pos= %d; n= %d => no content after separator", pos, n)
					return -1, -1, fmt.Errorf("no content after separator")
				}
				klog.V(4).Infof("pos= %d; n= %d => found", pos, n)
				return pos, pos + 4, nil
			}
		}
	}

	panic("should not reach here")
}
