package staticgen

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/bitsgofer/notebook/internal/blogcontent"
)

/*
import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	postDir := "testdata"
	postTemplate := "testdata/template/template.html"
	outDir, err := ioutil.TempDir(os.TempDir(), "staticgen")
	if err != nil {
		t.Fatalf("cannot get temp dir %q, err= %v", outDir, err)
	}

	if err := Generate(postDir, postTemplate, outDir); err != nil {
		t.Errorf("want Generate() to return no error, got= %v", err)
	}
}
*/

func TestAll(t *testing.T) {
	if err := Generate(
		"/home/mtong/workspace/src/github.com/bitsgofer/notebook/newPosts",
		"/home/mtong/workspace/src/github.com/bitsgofer/notebook/templates",
		"/home/mtong/workspace/src/github.com/bitsgofer/notebook/assets",
		"/home/mtong/workspace/src/github.com/bitsgofer/notebook/newPublicHTML",
	); err != nil {
		t.Fatalf("cannot generate assets; err= %q", err)
	}
}

func TestGenerateAssets(t *testing.T) {
	if err := generateAssets("/home/mtong/workspace/src/github.com/bitsgofer/notebook/assets", "/home/mtong/workspace/src/github.com/bitsgofer/notebook/newPublicHTML"); err != nil {
		t.Fatalf("cannot generate assets; err= %q", err)
	}
}

func TestRender(t *testing.T) {
	tmpls, err := parseTemplates("/home/mtong/workspace/src/github.com/bitsgofer/notebook/templates")
	if err != nil {
		t.Fatalf("cannot load templates; err= %q", err)
	}

	if err := generateHTML("/home/mtong/workspace/src/github.com/bitsgofer/notebook/newPosts", "/home/mtong/workspace/src/github.com/bitsgofer/notebook/newPublicHTML", tmpls); err != nil {
		t.Fatalf("cannot generate HTMLs; err= %q; err==nil: %t", err, err == nil)
	}
}

func TestParseTemplates(t *testing.T) {
	tmpls, err := parseTemplates("/home/mtong/workspace/src/github.com/bitsgofer/notebook/templates")
	if err != nil {
		t.Fatalf("cannot load templates; err= %q", err)
	}

	mustParseRFC3339 := func(str string) time.Time {
		v, err := time.Parse(time.RFC3339, str)
		if err != nil {
			panic(fmt.Sprintf("cannot parse %q into as time.RFC3339", str))
		}

		return v
	}

	articles := []blogcontent.Article{
		blogcontent.Article{
			ID:  "abc",
			URL: "/techincal-1",
			Metadata: blogcontent.Metadata{
				Title:     "Technical 1",
				Author:    blogcontent.User("pusheen"),
				WrittenAt: mustParseRFC3339("2020-07-25T23:00:00Z"),
				Tags:      []blogcontent.Tag{blogcontent.TagProgramming},
				Summary: `<p>blah <strong>blah</strong></p>
<ul>
<li>blah</li>
<li>blah</li>
</ul>
<p>blah blah blah</p>
`,
			},
			Content: `<h2 id="section-9">Section 9</h2>
<p>All things change in a dynamic environment. Your effort to remain what you are is what limits you.</p>
<ul>
<li>Tachikoma 1</li>
<li>Tachikoma 2</li>
</ul>
`,
		},
		blogcontent.Article{
			ID:  "abc",
			URL: "/techincal-2",
			Metadata: blogcontent.Metadata{
				Title:     "Technical 2",
				Author:    blogcontent.User("pusheen"),
				WrittenAt: mustParseRFC3339("2020-07-25T23:00:00Z"),
				Tags:      []blogcontent.Tag{blogcontent.TagProgramming},
				Summary: `<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse convallis placerat mi in imperdiet. Proin egestas urna consequat suscipit convallis. Fusce eleifend tortor tortor, sed ullamcorper mauris varius eu. Quisque eu leo dapibus tellus venenatis maximus. Nam diam libero, mattis sit amet tempor eu, dictum vel libero. Sed sed facilisis magna. Donec ac elementum est. Nulla nec urna neque.</p>

<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse convallis placerat mi in imperdiet. Proin egestas urna consequat suscipit convallis. Fusce eleifend tortor tortor, sed ullamcorper mauris varius eu. Quisque eu leo dapibus tellus venenatis maximus. Nam diam libero, mattis sit amet tempor eu, dictum vel libero. Sed sed facilisis magna. Donec ac elementum est. Nulla nec urna neque</p>.
`,
			},
			Content: `<h2 id="section-9">Section 9</h2>
<p>All things change in a dynamic environment. Your effort to remain what you are is what limits you.</p>
<ul>
<li>Tachikoma 1</li>
<li>Tachikoma 2</li>
</ul>
`,
		},
		blogcontent.Article{
			ID:  "f36b42dfe11ca4847b23fa6f42b53c30",
			URL: "/stand-alone-complex",
			Metadata: blogcontent.Metadata{
				Title:     "Stand Alone Complex",
				Author:    blogcontent.User("pusheen"),
				WrittenAt: mustParseRFC3339("2020-07-25T23:00:00Z"),
				Tags:      []blogcontent.Tag{blogcontent.TagProgramming},
				Summary:   `<p>blah blah blah blah blah</p>`,
			},
			Content: `<h2 id="section-9">Section 9</h2>
<p>All things change in a dynamic environment. Your effort to remain what you are is what limits you.</p>
<ul>
<li>Tachikoma 1</li>
<li>Tachikoma 2</li>
</ul>
`,
		},
	}

	var buf bytes.Buffer
	if err := tmpls.index.Execute(&buf, articles); err != nil {
		t.Fatalf("cannot execute template; err= %q", err)
	}

	fmt.Println("**********")
	fmt.Println(buf.String())
	fmt.Println("**********")

}
