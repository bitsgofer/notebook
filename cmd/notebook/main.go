package main

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/klog/v2"
)

var (
	generate            = kingpin.Command("generate", "generate static blog")
	generatePostDir     = generate.Flag("post-dir", "post directory").Default("posts").String()
	generateTemplateDir = generate.Flag("template-dir", "templates directory").Default("templates").String()
	generateHTMLDir     = generate.Flag("html-dir", "directory for resulting HTML/assets").Default("public_html").String()

	server                  = kingpin.Command("server", "serve static blog")
	serverEnableHTTPS       = server.Flag("https", "use HTTPS instead of HTTP").Default("true").Bool()
	serverLetsEncryptEmail  = server.Flag("lets-encrypt-email", "email used with Let's Encrypt").Default("admin@example.com").String()
	serverLetsEncryptDomain = server.Flag("lets-encrypt-domain", "comma-separated domains used with Let's Encrypt").Default("example.com,www.example.com").String()
)

func main() {
	klog.InitFlags(nil)

	switch kingpin.Parse() {
	case "generate":
	case "server":
	}

	fmt.Println("done")
}
