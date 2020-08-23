package blogcontent

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

var pandoc string

func init() {
	path, err := exec.LookPath("pandoc")
	if err != nil {
		if os.IsNotExist(err) {
			klog.Errorf("pandoc not found, please install (e.g: sudo apt-get install pandoc). Then make sure `which pandoc` works")
			os.Exit(1)
		}

		klog.Fatalf("cannot find pandoc; err= %q", err)
	}

	klog.V(4).Infof("pandoc's full path= %q", path)
	pandoc = path
}

// ToHTML converts markdown to HTMl5.
func ToHTML(b []byte) ([]byte, error) {
	f, err := ioutil.TempFile(os.TempDir(), "notebook-")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file; err= %w", err)
	}
	klog.V(4).Infof("temp file= %q", f.Name())

	if _, err := f.Write(b); err != nil {
		return nil, fmt.Errorf("cannot write to temp file; err= %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("cannot close temp file; err= %w", err)
	}
	klog.V(4).Infof("wrote markdown to temp file= %q", f.Name())
	defer os.Remove(f.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{
		"--from=markdown",
		"--to=html5",
		f.Name(),
	}
	cmd := exec.CommandContext(ctx, pandoc, args...)
	klog.V(4).Infof("%s %s", pandoc, strings.Join(args, " "))

	html5, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot run pandoc command; err= %w", err)
	}
	klog.V(4).Infof("html5= %q", html5)

	return html5, nil
}
