package post

import (
	"html/template"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func mustParseRFC3339ToUTC(str string) time.Time {
	v, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic("not formatted with RFC3339")
	}

	return v.UTC()
}

func callerPath() string {
	_, path, _, _ := runtime.Caller(0)
	abs, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		panic("cannot get current directory")
	}

	return abs
}

func TestNew(t *testing.T) {
	caller := callerPath()

	var testCases = []struct {
		name        string
		filename    string
		expectedErr error
		expected    *Post
	}{
		{
			name:        "happy path",
			filename:    "testdata/happy.md",
			expectedErr: nil,
			expected: &Post{
				Filename: caller + "/testdata/happy.md",
				Metadata: Metadata{
					Title:       "untitled",
					Slug:        "untitled",
					Author:      "mark",
					PublishedAt: mustParseRFC3339ToUTC("2018-07-21T08:00:00+08:00"),
					Sticky:      true,
					Tags:        []string{"group1", "tags2", "test"}, // sorted
				},
				HTML: template.HTML("<h1>h1</h1>\n\n<p>pppppp\npppp</p>\n\n<ul>\n<li>li</li>\n<li>li</li>\n</ul>\n\n<blockquote>\n<p>bq</p>\n</blockquote>\n\n<pre><code>pre\ncode\n</code></pre>\n"),
			},
		},
		{
			name:        "no metadata",
			filename:    "testdata/no_metadata.md",
			expectedErr: errors.Errorf("missing metadata section"),
			expected:    nil,
		},
		{
			name:        "missing separator",
			filename:    "testdata/missing_separator.md",
			expectedErr: errors.Errorf("missing metadata section"),
			expected:    nil,
		},
		{
			name:        "invalid metadata yaml",
			filename:    "testdata/invalid_metadata_yaml.md",
			expectedErr: errors.Errorf("yaml: unmarshal errors:\n  line 5: cannot unmarshal !!str `true - ...` into bool"),
			expected:    nil,
		},
		{
			name:        "invalid published at",
			filename:    "testdata/invalid_published_at.md",
			expectedErr: errors.Errorf(`parsing time "2018-07-21" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`),
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := New(tc.filename)
			switch {
			case tc.expectedErr == nil && err != nil:
				t.Errorf("want New() to return no error, got= %v", err)
			case tc.expectedErr != nil && err == nil:
				t.Errorf("want New() to return error= %v, got none", tc.expectedErr)
			case tc.expectedErr != nil && err != nil && errors.Cause(err).Error() != tc.expectedErr.Error():
				t.Errorf("want New() to return error= %v, got= %v", tc.expectedErr, err)
			}

			if want, got := tc.expected, p; !cmp.Equal(want, got) {
				t.Errorf("mismatched Post\n  want= %#v\n   got= %#v\n  diff= %v", want, got, cmp.Diff(want, got))
			}
		})
	}
}

func TestParentPath(t *testing.T) {
	p, err := New("testdata/happy.md")
	if err != nil {
		t.Fatalf("cannot create Post, err= %v", err)
	}

	if want, got := "2018/07/21", p.ParentPath(); want != got {
		t.Errorf("mismatched ParentPath(), want= %v, got= %v", want, got)
	}
}

func TestCanonicalPath(t *testing.T) {
	p, err := New("testdata/happy.md")
	if err != nil {
		t.Fatalf("cannot create Post, err= %v", err)
	}

	if want, got := "2018/07/21/untitled", p.CanonicalPath(); want != got {
		t.Errorf("mismatched CanonicalPath(), want= %v, got= %v", want, got)
	}
}

// func TestRender(t *testing.T) {
// 	tmpl := loadTemplate()
// 	var buf bytes.Buffer
// 	if err := p.Render(w, tmpl); err != nil {
// 		t.Errorf("expected no error rendering, got= %v", err)
// 	}
//
// 	expected, err := ioutil.ReadFile(renderdHTML)
// 	if want, got := expected, buf.String(); !cmp.Equal(want, got) {
// 		t.Errorf("mismatched post's rendered HTML\n  want= %v\n   got= %v\n  diff= %v", want, got, cmp.Diff(want, got))
// 	}
// }
//
// func TestRenderFile(t *testing.T) {
// 	tmp := loadTemplate()
//
// 	if err := p.RenderFile(tempFile, tmpl); err != nil {
// 		t.Errorf("expected no error rendering, got= %v", err)
// 	}
//
// 	actual, err := ioutil.ReadFile(tempFile)
// 	expected, err := ioutil.ReadFile(renderdHTML)
// 	if want, got := expected, actual; !cmp.Equal(want, got) {
// 		t.Errorf("mismatched post's rendered HTML\n  want= %v\n   got= %v\n  diff= %v", want, got, cmp.Diff(want, got))
// 	}
// }
