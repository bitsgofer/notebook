# notebook

![build status](https://travis-ci.org/exklamationmark/notebook.svg?branch=master)

DIY blogging utility:

To use:

- Download `minify`: `go get github.com/tdewolff/minify/cmd/minify`
- Download [prism](https://prismjs.com/download.html#themes=prism-okaidia&languages=markup+clike+ada+c+asciidoc+asm6502+bash+cpp+clojure+ruby+d+dart+diff+docker+erlang+go+graphql+http+hpkp+java+json+julia+latex+markdown+lisp+lua+nginx+ocaml+pascal+perl+sql+protobuf+python+q+r+rust+scheme+smalltalk+yaml&plugins=line-numbers+command-line)
- Download [mini.css](https://github.com/Chalarangelo/mini.css/releases)
- Clone repo & build (e.g `make build`)
- Posts are written to `./posts`
- Run `./build/notebook generate`
- Run `./build/notebook server`
