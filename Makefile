_pkg=github.com/exklamationmark/notebook

PKG=./...
RUN=
ifneq ("$(RUN)","")
	_go_test_run_flag=-run $(RUN)
endif
GLOG=--stderrthreshold=FATAL

all: test build gen
.PHONY: all

test:
	go test -cover -v -race $(_pkg)/$(PKG) ${_go_test_run_flag} $(GLOG)
	go vet $(_pkg)/$(PKG)
.PHONY: test

build:
	mkdir -p build
	go build -o  build/notebook ${_pkg}/cmd/notebook
.PHONY: build

gen.assets:
	minify assets/*.css > public_html/builtin.css
	minify assets/*.js > public_html/builtin.js
	cp assets/favicon.ico public_html/favicon.ico
.PHONY: gen.assets

gen: clean gen.assets
	./build/notebook generate --post.dir=posts/ --html.dir=public_html/ --post.template=template.html
.PHONY: gen

serve.local:
	sudo ./build/notebook serve --html.dir=public_html/
.PHONY: serve.local

run: build gen serve.local
.PHONY: run

clean:
	rm -rf public_html/*/
	rm -f public_html/*.css
	rm -f public_html/*.js
	rm -f public_html/*.ico
.PHONY: clean
