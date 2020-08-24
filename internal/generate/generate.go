package generate

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"k8s.io/klog/v2"

	"github.com/bitsgofer/notebook/internal/blogcontent"
)

var minify string

func init() {
	path, err := exec.LookPath("minify")
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			klog.Errorf("minify not found, please install (e.g: go install github.com/tdewolff/minify/cmd/minify). Then make sure `which minify` works")
			os.Exit(1)
		}

		klog.Fatalf("cannot find minify; err= %q", err)
	}

	klog.V(4).Infof("minify's full path= %q", path)
	minify = path
}

// blogTemplates templates for rendering different type of pages
// (e.g: single-article, index, etc).
type blogTemplates struct {
	index         *template.Template
	singleArticle *template.Template
	errorPage     *template.Template
}

// loadTemplates combines files in <templateDir> into blogTemplates.
// It assumes template files are using default names (e.g: "index.html", "menu.html", etc).
func loadTemplates(templateDir string) (*blogTemplates, error) {
	var loadErrors *multierror.Error

	// load abstracts loading files to create a template.
	// The first template must be able to render the whole page.
	// Others must only contain nested templates.
	// Errors are accumulated to <loadErrors>.
	load := func(names ...string) *template.Template {
		var paths []string
		for _, name := range names {
			paths = append(paths, templateDir+"/"+name)

		}
		klog.V(4).Infof("loading template files: %v", paths)

		tmpl, err := template.ParseFiles(paths...)
		if err != nil {
			loadErrors = multierror.Append(loadErrors,
				fmt.Errorf("cannot load template files= %v; err= %w", paths, err))
		}

		return tmpl
	}

	// templateFilesFor returns a list of template files to render page.
	// The first arg must be the main template that renders the full HTML.
	// Core templates will be appended automatically.
	templateFilesFor := func(mainTemplate string, nestedTemplates ...string) []string {
		var all []string
		all = append(all, mainTemplate)
		all = append(all, nestedTemplates...)
		all = append(all, "head.html")
		all = append(all, "script.html")
		all = append(all, "menu.html")
		all = append(all, "footer.html")

		return all
	}

	templates := blogTemplates{
		index:         load(templateFilesFor("index.html", "articleSummary.html")...),
		singleArticle: load(templateFilesFor("single-article.html", "article.html")...),
		errorPage:     load(templateFilesFor("error.html")...),
	}
	if loadErrors != nil {
		return nil, loadErrors
	}

	return &templates, nil
}

// generateHTML walks the <postDir> and convert each (pandoc) markdown files
// into HTML5.
func generateHTML(postDir, outDir string, tmpls *blogTemplates, cssBytes, jsBytes []byte) error {
	var renderErrors *multierror.Error

	// stringify the CSS and JS bytes for rendering later
	css := template.CSS(cssBytes)
	js := template.JS(jsBytes)

	// create folder for posts (/posts/*.html)
	postsOutPath := outDir + "/posts"
	if err := os.MkdirAll(postsOutPath, os.ModeDir|0755); err != nil {
		return fmt.Errorf("cannot create posts dir; err= %w", err)
	}

	// renderToFile abstracts rendering a template into a file
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
	// walkFn process each file within <postDir>, rendering markdown articles.
	// It also accumulates these articles into <articles> to generate
	// the index page later.
	walkFn := func(path string, stat os.FileInfo, walkErr error) error {
		klog.V(4).Infof("processing %q; walkErr= %q; stat= %v", path, walkErr, stat)
		if walkErr != nil {
			klog.V(4).Infof("encountered walkErr= %q", walkErr)
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("filepath.Walk() gave error for %q; err= %w", path, walkErr))
			return nil
		}

		// only process .md files
		if stat.IsDir() {
			klog.V(4).Infof("%q is a directory", path)
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			klog.V(4).Infof("%q is not a markdown file (.md)", path)
			return nil
		}

		// parse article from file
		b, err := ioutil.ReadFile(path)
		if err != nil {
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("cannot read %q; err= %w", path, err))
			return nil
		}
		article, err := blogcontent.ParseArticle(bytes.NewBuffer(b))
		if err != nil {
			renderErrors = multierror.Append(renderErrors,
				fmt.Errorf("cannot parse article from %q; err= %w", path, err))
			return nil
		}
		klog.V(4).Infof("parsed article from %q", path)

		// render to <postsOutPath>/<.Metadata.FileName>
		outPath := fmt.Sprintf("%s/%s", postsOutPath, article.FileName)
		// wrap an Article together with CSS and JS for rendering
		type wrapper struct {
			Article *blogcontent.Article
			CSS     template.CSS
			JS      template.JS
		}
		wrapped := wrapper{
			Article: article,
			CSS:     css,
			JS:      js,
		}
		if err := renderToFile(tmpls.singleArticle, wrapped, outPath); err != nil {
			renderErrors = multierror.Append(renderErrors, err)
			return nil
		}
		klog.V(2).Infof("rendered %q into %q", path, outPath)

		// accumulate articles for index page
		articles = append(articles, article)

		return nil
	}

	// create html pages for each articles + index page
	if err := filepath.Walk(postDir, walkFn); err != nil {
		return fmt.Errorf("cannot render artiles; err= %w", err)
	}
	indexPath := fmt.Sprintf("%s/%s", outDir, "index.html")
	// wrap multiple Articles together with CSS and JS for rendering
	type wrapper struct {
		Articles []*blogcontent.Article
		CSS      template.CSS
		JS       template.JS
	}
	wrapped := wrapper{
		Articles: articles,
		CSS:      css,
		JS:       js,
	}

	if err := renderToFile(tmpls.index, wrapped, indexPath); err != nil {
		renderErrors = multierror.Append(renderErrors,
			fmt.Errorf("cannot render index page; err= %w", err))
	}
	klog.V(2).Infof("rendered index page into %q", indexPath)

	// create error page
	errorPagePath := outDir + "/error.html"
	noContent := wrapper{
		CSS: css,
		JS:  js,
	}
	if err := renderToFile(tmpls.errorPage, noContent, errorPagePath); err != nil {
		renderErrors = multierror.Append(renderErrors,
			fmt.Errorf("cannot render error page; err= %w", err))
	}
	klog.V(2).Infof("rendered error page into %q", errorPagePath)

	// need explicity check, otherwise (*multierror.Error)(nil) will be
	// converted to a non-nil error
	if renderErrors != nil {
		return renderErrors
	}
	return nil
}

// generateAssets prepares non-HTML content (CSS, JS, etc) for the blog.
// For CSS and JS, it will combine all files into one minified file.
//
// We also return the minified CSS and JS bytes to be included in each HTML file,
// instead of loading them separately. This will speed up page load.
// REF: https://web.dev/render-blocking-resources
func generateAssets(assetsDir, outDir string) (cssBytes, jsBytes []byte, err error) {
	// minifyToFile combines and minify multiple CSS/JS files into one file.
	minifyToFile := func(outPath string, files ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, minify, files...)
		klog.V(4).Infof("%s %s", minify, strings.Join(files, " "))

		// run minify command, result is in stdout
		stdout, err := cmd.Output()
		if err != nil {
			exitErr, _ := err.(*exec.ExitError)
			klog.V(4).Infof("minify failed; stderr= %q", exitErr.Stderr)
			return fmt.Errorf("cannot execute minfy command= %v; err= %w", cmd, err)
		}
		klog.V(4).Infof("executed %v", cmd)

		// write minified content <outPath>
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
		klog.V(2).Infof("minified {%s} into %q", strings.Join(files, ", "), outPath)

		return nil
	}

	// find all CSS, JS and images files in <assetsDir>
	var cssFiles, jsFiles, imageFiles []string
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
		switch strings.ToLower(filepath.Ext(path)) {
		case ".css":
			cssFiles = append(cssFiles, path)
			klog.V(4).Infof("%q is a CSS file", path)
		case ".js":
			jsFiles = append(jsFiles, path)
			klog.V(4).Infof("%q a JS file", path)
		case ".png", ".jpg", ".jpeg", ".gif":
			imageFiles = append(imageFiles, path)
		default:
		}

		return nil
	}
	if err := filepath.Walk(assetsDir, walkFn); err != nil {
		return nil, nil, fmt.Errorf("cannot find CSS and JS files; err= %w", err)
	}

	// create minified CSS + JS + favicon
	assetsOutPath := outDir + "/assets"
	if err := os.MkdirAll(assetsOutPath, os.ModeDir|0755); err != nil {
		return nil, nil, fmt.Errorf("cannot create assets dir; err= %w", err)
	}
	cssOutPath := outDir + "/assets/notebook.css"
	if err := minifyToFile(cssOutPath, cssFiles...); err != nil {
		return nil, nil, fmt.Errorf("cannot minfy CSS files; err= %w", err)
	}
	jsOutPath := outDir + "/assets/notebook.js"
	if err := minifyToFile(jsOutPath, jsFiles...); err != nil {
		return nil, nil, fmt.Errorf("cannot minify JS files; err= %w", err)
	}
	faviconOutPath := outDir + "/assets/favicon.ico"
	if err := exec.Command("cp", assetsDir+"/favicon.ico", faviconOutPath).Run(); err != nil {
		return nil, nil, fmt.Errorf("cannot create favicon; err= %w", err)
	}
	klog.V(2).Infof("copied favicon to %q", faviconOutPath)
	errorPageImgOutPath := outDir + "/assets/error.jpg"
	if err := exec.Command("cp", assetsDir+"/error.jpg", errorPageImgOutPath).Run(); err != nil {
		return nil, nil, fmt.Errorf("cannot create error image; err= %w", err)
	}
	klog.V(2).Infof("copied error page image to %q", errorPageImgOutPath)

	// read the minified CSS and JS bytes
	cssBytes, err = ioutil.ReadFile(cssOutPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read minified CSS file; err= %w", err)
	}
	jsBytes, err = ioutil.ReadFile(jsOutPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read minified JS file; err= %w", err)
	}

	return cssBytes, jsBytes, nil
}

// Config contains configurations to generate static blog.
type Config struct {
	PostsDir     string
	TemplatesDir string
	AssetsDir    string
	OutputDir    string
}

// Generate creates a clean static blog.
func Generate(cfg *Config) error {
	if err := canonizeConfigAndValidate(cfg); err != nil {
		return fmt.Errorf("invalid config; err= %w", err)
	}

	if err := os.RemoveAll(cfg.OutputDir); err != nil {
		return fmt.Errorf("cannot remove %q; err= %w", cfg.OutputDir, err)
	}
	if err := os.MkdirAll(cfg.OutputDir, os.ModeDir|0755); err != nil {
		return fmt.Errorf("cannot create %q; err= %w", cfg.OutputDir, err)
	}

	templates, err := loadTemplates(cfg.TemplatesDir)
	if err != nil {
		return fmt.Errorf("cannot load templates; err= %w", err)
	}

	cssBytes, jsBytes, err := generateAssets(cfg.AssetsDir, cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("cannot generate assets; err= %w", err)
	}

	if err := generateHTML(cfg.PostsDir, cfg.OutputDir, templates, cssBytes, jsBytes); err != nil {
		return fmt.Errorf("cannot generate articles; err= %w", err)
	}

	return nil
}

func canonizeConfigAndValidate(cfg *Config) error {
	postsDirAbsPath, err := filepath.Abs(cfg.PostsDir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path to .PostsDir; err= %w", err)
	}
	cfg.PostsDir = postsDirAbsPath

	templatesDirAbsPath, err := filepath.Abs(cfg.TemplatesDir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path to .PostsDir; err= %w", err)
	}
	cfg.TemplatesDir = templatesDirAbsPath

	assetsDirAbsPath, err := filepath.Abs(cfg.AssetsDir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path to .PostsDir; err= %w", err)
	}
	cfg.AssetsDir = assetsDirAbsPath

	outputDirAbsPath, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path to .PostsDir; err= %w", err)
	}
	cfg.OutputDir = outputDirAbsPath

	return nil
}
