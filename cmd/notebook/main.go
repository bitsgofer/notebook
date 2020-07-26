package main

import (
	"fmt"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/klog/v2"

	"github.com/bitsgofer/notebook/internal/generate"
)

var (
	// generic flags
	klogV = kingpin.Flag("v", "Enable V-leveled logging at the specified level.").Default("0").String()

	// generate: create HTML from pandoc Markdown files
	generateCmd          = kingpin.Command("generate", "Generate static blog.")
	generatePostsDir     = generateCmd.Flag("post-dir", "Post directory.").Default(fullPath("newPosts")).String()
	generateTemplatesDir = generateCmd.Flag("template-dir", "Templates directory.").Default(fullPath("templates")).String()
	generateAssetsDir    = generateCmd.Flag("assets-dir", "Assets directory.").Default(fullPath("assets")).String()
	generateOutputDir    = generateCmd.Flag("out-dir", "Directory for resulting HTML and assets.").Default(fullPath("newPublicHTML")).String()

	// server: run HTTP/HTTPS server for the static pages
	serverCmd                = kingpin.Command("server", "Serve static blog")
	serverEnableHTTPS        = serverCmd.Flag("https", "Use HTTPS instead of HTTP.").Default("true").Bool()
	serverLetsEncryptEmail   = serverCmd.Flag("admin-email", "Email used with Let's Encrypt.").Default("admin@example.com").String()
	serverLetsEncryptDomains = serverCmd.Flag("domains", "(Multiple) domains used with Let's Encrypt.").Default("example.com", "www.example.com").Strings()
)

func fullPath(dir string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		klog.Fatalf("cannot get current working directory; err= %q", err)
	}

	return fmt.Sprintf("%s/%s", currentDir, dir)
}

func main() {
	cmd := kingpin.Parse()

	// Some hack to get back a **some** functionality of klog.
	// klog's stderrThreshold will always be ERROR
	// Obviously klog doesn't fit with kinpin :P
	klogVLevel := klog.Level(0)
	(&klogVLevel).Set(*klogV)

	switch cmd {
	case "generate":
		cfg := generate.Config{
			PostsDir:     *generatePostsDir,
			TemplatesDir: *generateTemplatesDir,
			AssetsDir:    *generateAssetsDir,
			OutputDir:    *generateOutputDir,
		}
		klog.V(2).Infof("running: generate with config= %#v", cfg)

		if err := generate.Generate(cfg); err != nil {
			klog.Fatalf("cannot generate blog; err= %q", err)
		}

	case "server":
		klog.V(2).Infof("running: server; https= %t; email= %q; domains= %v", *serverEnableHTTPS, *serverLetsEncryptEmail, *serverLetsEncryptDomains)
	}

	fmt.Println("done")
}
