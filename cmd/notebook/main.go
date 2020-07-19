package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/bitsgofer/notebook/internal/blog"
	"github.com/bitsgofer/notebook/internal/middlewares/redirect"
	"github.com/bitsgofer/notebook/internal/staticgen"
)

type domainsFlag []string

func (d *domainsFlag) String() string {
	return fmt.Sprint(*d)
}

func (d *domainsFlag) Set(str string) error {
	for _, domain := range strings.Split(str, ",") {
		*d = append(*d, domain)
	}

	return nil
}

type config struct {
	htmlDir string

	// 	generate
	postDir      string
	postTemplate string

	// server
	adminEmail   string
	domains      domainsFlag
	production   bool
	redirections redirect.Redirections
}

func main() {
	var c config

	a := kingpin.New(filepath.Base(os.Args[0]), "notebook application")
	a.HelpFlag.Short('h')

	a.Flag("html.dir", "output html directory").Default("public_html").
		StringVar(&c.htmlDir)

	gen := a.Command("generate", "generate HTML from posts")

	gen.Flag("post.dir", "post directory").Default("posts").
		StringVar(&c.postDir)

	gen.Flag("post.template", "post template file").Default("template.html").
		StringVar(&c.postTemplate)

	server := a.Command("serve", "run blog server")

	server.Flag("admin.email", "admin email for Let's Encrypt").Default("admin@example.com").
		StringVar(&c.adminEmail)

	server.Flag("domains", "domains to obtain TLS cert").Default("example.com").
		SetValue(&c.domains)

	server.Flag("production", "production mode (enable HTTPS)").Default("false").
		BoolVar(&c.production)

	server.Flag("redirect", "redirection options").Default("").
		SetValue(&c.redirections)

	// ----------------------------------------

	cmd, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}
	fmt.Printf("%#v\n", cmd)

	switch cmd {
	case "generate":
		if err := staticgen.Generate(c.postDir, c.postTemplate, c.htmlDir); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error generating html"))
			os.Exit(1)
		}
	case "serve":
		fmt.Printf("%#v\n", c)
		srv, err := blog.New(c.htmlDir, c.adminEmail, c.domains, blog.Redirect(c.redirections))
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Cannot create server"))
			os.Exit(1)
		}

		if !c.production {
			runInDev(srv)
			return
		}
		runInProd(srv)
	}
}

func runInProd(srv *blog.Server) {
	go func() {
		httpSrv := &http.Server{
			Addr:    ":80",
			Handler: srv.HTTPRedirectHandler(),
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrapf(err, "HTTP server failed"))
			os.Exit(1)
		}
	}()

	blogSrv := &http.Server{
		Addr:      ":443",
		Handler:   srv.BlogHandler(),
		TLSConfig: srv.TLSConfig(),
	}
	if err := blogSrv.ListenAndServeTLS("", ""); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Blog server failed"))
		os.Exit(1)
	}
}

func runInDev(srv *blog.Server) {
	blogSrv := &http.Server{
		Addr:    ":80",
		Handler: srv.BlogHandler(),
	}
	if err := blogSrv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "blog server failed"))
		os.Exit(1)
	}

}
