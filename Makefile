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

gen: build
	./build/notebook generate --post.dir=posts/ --html.dir=public_html/ --post.template=template.html
.PHONY: gen

serve.local: build
serve.local: gen
	sudo ./build/notebook serve --html.dir=public_html/
.PHONY: run
