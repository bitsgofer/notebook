package blogcontent

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	blogPostEmpty              = ``
	blogPostWithMetaAndContent = `metadata:
  key-1: value--1
  key-2: ---value-2

----

content: blah blah
----
blah blah`  // no usable metadata
	blogPostMetadataOnly = `metadata:
  key: value
----`
	blogPostNoSeparator = `metadata:
  key: value
-- --
blah blah`
	blogPostValid = `
title: Stand Alone Complex
written_at: 2020-07-25T23:00:00Z
author: pusheen
tags:
- programming
summary: |
  blah **blah**

  - blah
  - blah

  blah blah blah
----

Stand Alone Complex
===================

Section 9
---------

All things change in a dynamic environment. Your effort to remain what you are is what limits you.

- Tachikoma 1
- Tachikoma 2`
)

func TestToHTML(t *testing.T) {
	raw := `
Stand Alone Complex
===================

Section 9
---------

All things change in a dynamic environment. Your effort to remain what you are is what limits you.

- Tachikoma 1
- Tachikoma 2`
	wantHTML := `<h1 id="stand-alone-complex">Stand Alone Complex</h1>
<h2 id="section-9">Section 9</h2>
<p>All things change in a dynamic environment. Your effort to remain what you are is what limits you.</p>
<ul>
<li>Tachikoma 1</li>
<li>Tachikoma 2</li>
</ul>
`

	html, err := ToHTML([]byte(raw))
	if err != nil {
		t.Fatalf("want no error; got= %q", err)
	}

	if want, got := wantHTML, string(html); !cmp.Equal(want, got) {
		t.Fatalf("wrong HTML output; diff= %s", cmp.Diff(want, got))
	}
}

func mustParseRFC3339(str string) time.Time {
	v, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(fmt.Sprintf("cannot parse %q into as time.RFC3339", str))
	}

	return v
}

func TestParseArticle(t *testing.T) {
	cmpOpts := []cmp.Option{
		cmp.AllowUnexported(Article{}),
	}

	var testCases = map[string]struct {
		raw         string
		isErr       bool
		wantArticle *Article
	}{
		"empty": {
			raw:   blogPostEmpty,
			isErr: true,
		},
		"no-separtor": {
			raw:   blogPostNoSeparator,
			isErr: true,
		},
		"metadata-only": {
			raw:   blogPostMetadataOnly,
			isErr: true,
		},
		"valid": {
			raw: blogPostValid,
			wantArticle: &Article{
				ID:       "f36b42dfe11ca4847b23fa6f42b53c30",
				URL:      "/stand-alone-complex",
				FileName: "stand-alone-complex.html",
				Metadata: Metadata{
					Title:     "Stand Alone Complex",
					Author:    User("pusheen"),
					WrittenAt: mustParseRFC3339("2020-07-25T23:00:00Z"),
					Tags:      []Tag{TagProgramming},
					RawSummary: `blah **blah**

- blah
- blah

blah blah blah
`,
					Summary: `<p>blah <strong>blah</strong></p>
<ul>
<li>blah</li>
<li>blah</li>
</ul>
<p>blah blah blah</p>
`,
				},
				content: []byte(`

Stand Alone Complex
===================

Section 9
---------

All things change in a dynamic environment. Your effort to remain what you are is what limits you.

- Tachikoma 1
- Tachikoma 2`),
				Content: `<h1 id="stand-alone-complex">Stand Alone Complex</h1>
<h2 id="section-9">Section 9</h2>
<p>All things change in a dynamic environment. Your effort to remain what you are is what limits you.</p>
<ul>
<li>Tachikoma 1</li>
<li>Tachikoma 2</li>
</ul>
`,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.raw)
			article, err := ParseArticle(buf)

			switch {
			case tc.isErr && err == nil: // bad
				t.Fatalf("want error, got none")
			case tc.isErr && err != nil: // okay
				return
			case !tc.isErr && err != nil: // bad
				t.Fatalf("want no error, got err= %q", err)
			default: // !tc.isErr && err == nil: // okay
			}

			if want, got := tc.wantArticle, article; !cmp.Equal(want, got, cmpOpts...) {
				t.Fatalf("diff= %s", cmp.Diff(want, got, cmpOpts...))
			}
		})
	}
}

func TestSplitBySeparator(t *testing.T) {
	var testCases = map[string]struct {
		raw              string
		isErr            bool
		wantMetaEnd      int
		wantContentStart int
	}{
		"empty": {
			raw:   blogPostEmpty,
			isErr: true,
		},
		"no-separtor": {
			raw:   blogPostNoSeparator,
			isErr: true,
		},
		"metadata-only": {
			raw:   blogPostMetadataOnly,
			isErr: true,
		},
		"valid": {
			raw:              blogPostWithMetaAndContent,
			wantMetaEnd:      49,
			wantContentStart: 53,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			metaEnd, contentStart, err := splitBySeparator([]byte(tc.raw))

			switch {
			case tc.isErr && err == nil: // bad
				t.Fatalf("want error, got none")
			case tc.isErr && err != nil: // okay
				return
			case !tc.isErr && err != nil: // bad
				t.Fatalf("want no error, got err= %q", err)
			default: // !tc.isErr && err == nil: // okay
			}

			// t.Logf("raw:\n***\n%s\n***\nmetadata:\n***\n%s\n***\ncontent:\n***\n%s\n***\n",
			// 	tc.raw, string(tc.raw[:metaEnd]), string(tc.raw[contentStart:]))
			if tc.wantMetaEnd != metaEnd && tc.wantContentStart != contentStart {
				t.Fatalf("want= (metaEnd= %d, contentStart= %d); got= (metaEnd= %d, contentStart= %d",
					tc.wantMetaEnd, tc.wantContentStart, metaEnd, contentStart)
			}
		})
	}
}
