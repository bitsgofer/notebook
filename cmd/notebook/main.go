package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/exklamationmark/notebook/internal/blogserver"
	"github.com/exklamationmark/notebook/internal/staticgen"
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
	httpAddr   string
	httpsAddr  string
	adminEmail string
	domains    domainsFlag
	production bool
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

	server.Flag("http.addr", "HTTP listen address (e.g: ':80')").Default(":80").
		StringVar(&c.httpAddr)

	server.Flag("https.addr", "HTTPS listen address (e.g: ':443')").Default(":443").
		StringVar(&c.httpsAddr)

	server.Flag("admin.email", "admin email for Let's Encrypt").Default("admin@example.com").
		StringVar(&c.adminEmail)

	server.Flag("domains", "domains to obtain TLS cert").Default("example.com").
		SetValue(&c.domains)

	server.Flag("production", "production mode (enable HTTPS)").Default("false").
		BoolVar(&c.production)

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
		srv, err := blogserver.New(c.htmlDir, c.adminEmail, c.domains...)
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Cannot create server"))
			os.Exit(1)
		}

		go func() {
			if c.production {
				if err := srv.HTTPSServer().ListenAndServeTLS("", ""); err != nil {
					fmt.Fprintln(os.Stderr, errors.Wrapf(err, "HTTPS server failed"))
					os.Exit(1)
				}
			}
		}()

		if err := srv.HTTPServer().ListenAndServe(); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrapf(err, "HTTP server failed"))
			os.Exit(1)
		}

	}
}
