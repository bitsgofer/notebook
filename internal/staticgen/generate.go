package staticgen

import (
	/*
		"bytes"
		"fmt"
		"html/template"
		"os"
		"path/filepath"
		"sort"
		"strings"

		"github.com/bitsgofer/notebook/internal/post"
		"github.com/exklamationmark/glog"
		"github.com/pkg/errors"
	*/

	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"k8s.io/klog/v2"

	"github.com/bitsgofer/notebook/internal/blogcontent"
)

type blogTemplates struct {
	index         *template.Template
	singleArticle *template.Template
}

// parseTemplates load template files for rendering single-article, index pages.
func parseTemplates(templateDir string) (*blogTemplates, error) {
	var loadErrors *multierror.Error

	// load a set of template files within templateDir
	load := func(names ...string) *template.Template {
		var paths []string
		for _, name := range names {
			paths = append(paths, templateDir+"/"+name)

		}
		tmpl, err := template.ParseFiles(paths...)
		if err != nil {
			loadErrors = multierror.Append(loadErrors,
				fmt.Errorf("cannot load template files= %v; err= %w", paths, err))
		}

		return tmpl
	}

	templates := blogTemplates{
		index:         load("index.html", "articleSummary.html", "head.html", "script.html", "menu.html", "footer.html"),
		singleArticle: load("single-article.html", "article.html", "head.html", "script.html", "menu.html", "footer.html"),
	}
	if loadErrors != nil {
		return nil, loadErrors
	}

	return &templates, nil
}

// generateHTML walks the <postDir> and render each file
func generateHTML(postDir, outDir string, tmpls *blogTemplates) error {
	var renderErrors *multierror.Error

	// renderToFile abstracts rendering a template + writing result to file
	renderToFile := func(tmpl *template.Template, data interface{}, outPath string) error {
		w, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("cannot create %q to write; err= %w", outPath, err)
		}
		klog.V(4).Infof("created %q to write", outPath)

		if err := tmpl.Execute(w, data); err != nil {
			return fmt.Errorf("cannot render article to %q; err= %w", outPath, err)
		}
		klog.V(4).Infof("rendered article to %q", outPath)

		if err := w.Close(); err != nil {
			return fmt.Errorf("cannot close (save) %q; err= %w", outPath, err)
		}
		klog.V(4).Infof("saved %q", outPath)

		return nil
	}

	var articles []*blogcontent.Article
	// walkFn process each file/directory within <postDir>, including itself.
	// It also accumulates parsed articles into <articles>.
	walkFn := func(path string, stat os.FileInfo, walkErr error) error {
		klog.V(4).Infof("processing %q; walkErr= %q; stat= %v", path, walkErr, stat)
		if walkErr != nil {
			klog.V(4).Infof("encountered walkErr= %q", walkErr)
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("filepath.Walk() gave error for %q; err= %w", path, walkErr))
			return nil
		}

		if stat.IsDir() { // don't process directory
			klog.V(4).Infof("%q is a directory", path)
			return nil
		}
		if filepath.Ext(path) != ".md" {
			klog.V(4).Infof("%q is not a markdown file (.md)", path)
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("cannot open %q to read; err= %w", path, err))
			return nil
		}
		defer f.Close()
		klog.V(4).Infof("opened %q to read", path)

		article, err := blogcontent.ParseArticle(f)
		if err != nil {
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("cannot parse article from %q; err= %w", path, err))
			return nil
		}
		klog.V(4).Infof("parsed article from %q", path)

		outPath := fmt.Sprintf("%s/%s", outDir, article.FileName)
		if err := renderToFile(tmpls.singleArticle, article, outPath); err != nil {
			renderErrors = multierror.Append(renderErrors, err)
			return nil
		}

		articles = append(articles, article)
		return nil
	}

	// render articles + index
	if err := filepath.Walk(postDir, walkFn); err != nil {
		return fmt.Errorf("cannot render artiles; err= %w", err)
	}
	indexPath := fmt.Sprintf("%s/%s", outDir, "index.html")
	if err := renderToFile(tmpls.index, articles, indexPath); err != nil {
		renderErrors = multierror.Append(renderErrors, err)
	}

	// need explicity check, otherwise (*multierror.Error)(nil) will be
	// converted to a non-nil error
	if renderErrors != nil {
		return renderErrors
	}
	return nil
}

// generateAssets prepares non-HTML (CSS, JS, etc) content for the blog.
// For CSS and JS, it will combine all contents into one minified file.
func generateAssets(assetsDir, outDir string) error {
	pipeCmdOutputToFile := func(outPath string, command string, args ...string) error {
		cmdPath, err := exec.LookPath(command)
		if err != nil {
			return fmt.Errorf("cannot find full path for %q; err= %q", command, err)
		}
		klog.V(4).Infof("%q's full path= %q", command, cmdPath)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, cmdPath, args...)
		klog.V(4).Infof("%s %s", cmdPath, strings.Join(args, " "))

		stdout, err := cmd.Output()
		if err != nil {
			klog.V(4).Infof("err= %#v", err)
			exitErr, _ := err.(*exec.ExitError)
			klog.V(4).Infof("stderr= %s", string(exitErr.Stderr))
			return fmt.Errorf("cannot execute command %v; err= %w", cmd, err)
		}
		klog.V(4).Infof("executed %v", cmd)

		w, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("cannot create %q to write; err= %w", outPath, err)
		}
		klog.V(4).Infof("created %q to write", outPath)

		if _, err := w.Write(stdout); err != nil {
			return fmt.Errorf("cannot write to %q; err= %w", outPath, err)
		}
		klog.V(4).Infof("wrote to %q", outPath)

		if err := w.Close(); err != nil {
			return fmt.Errorf("cannot close (save) %q; err= %w", outPath, err)
		}
		klog.V(4).Infof("saved %q", outPath)

		return nil
	}

	// find all CSS and JS files in <assetsDir>
	var cssFiles, jsFiles []string
	walkFn := func(path string, stat os.FileInfo, walkErr error) error {
		klog.V(4).Infof("processing %q; walkErr= %q; stat= %v", path, walkErr, stat)
		if walkErr != nil {
			klog.V(4).Infof("encountered walkErr= %q", walkErr)
			return nil
		}

		if stat.IsDir() { // don't process directory
			klog.V(4).Infof("%q is a directory", path)
			return nil
		}
		switch filepath.Ext(path) {
		case ".css":
			cssFiles = append(cssFiles, path)
			klog.V(4).Infof("%q is a CSS file", path)
		case ".js":
			jsFiles = append(jsFiles, path)
			klog.V(4).Infof("%q a JS file", path)
		default:
		}

		return nil
	}
	if err := filepath.Walk(assetsDir, walkFn); err != nil {
		return fmt.Errorf("cannot find CSS and JS files; err= %w", err)
	}

	err := pipeCmdOutputToFile(outDir+"/notebook.css",
		"minify", cssFiles...)
	if err != nil {
		return fmt.Errorf("cannot minfy CSS files; err= %w", err)
	}
	err = pipeCmdOutputToFile(outDir+"/notebook.js",
		"minify", jsFiles...)
	if err != nil {
		return fmt.Errorf("cannot minify JS files; err= %w", err)
	}
	err = exec.Command("cp", assetsDir+"/favicon.ico", outDir+"/favicon.ico").Run()
	if err != nil {
		return fmt.Errorf("cannot prepare favicon; err= %w", err)
	}

	return nil
}

func Generate(postDir, templateDir, assetsDir, outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("cannot remove %q; err= %w", outDir, err)
	}
	if err := os.MkdirAll(outDir, os.ModeDir|0755); err != nil {
		return fmt.Errorf("cannot create %q; err= %w", outDir, err)
	}

	templates, err := parseTemplates(templateDir)
	if err != nil {
		return fmt.Errorf("cannot load templates; err= %w", err)
	}

	if err := generateHTML(postDir, outDir, templates); err != nil {
		return fmt.Errorf("cannot generate articles; err= %w", err)
	}

	if err := generateAssets(assetsDir, outDir); err != nil {
		return fmt.Errorf("cannot generate assets; err= %w", err)
	}

	return nil
}
