package blogcontent

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// Parser splits a blog post file into the medatadata and blog content.
type Parser struct {
}

// ParseArticle parses a file into an Article.
func ParseArticle(r io.Reader) (*Article, error) {
	metadata, content, err := readFile(r)
	if err != nil {
		return nil, err
	}

	article := Article{
		Content: content,
	}

	// unmarshal metadata
	buf := bytes.NewBuffer(metadata)
	dec := yaml.NewDecoder(buf)
	dec.KnownFields(true)
	if err := dec.Decode(&article.Metadata); err != nil {
		return nil, fmt.Errorf("cannot unmarshal metadata (yaml); err= %w", err)
	}

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
