package staticgen

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
