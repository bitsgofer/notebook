_pkg=github.com/exklamationmark/notebook

PKG=./...
RUN=
ifneq ("$(RUN)","")
	_go_test_run_flag=-run $(RUN)
endif
GLOG=
test:
	go test -cover -v -race $(_pkg)/$(PKG) ${_go_test_run_flag} $(GLOG)
	go vet $(_pkg)/$(PKG)
.PHONY: test

build:
	mkdir -p build
	go build -o  build/static-gen ${_pkg}/cmd/static-gen
	go build -o  build/blog-server ${_pkg}/cmd/blog-server
.PHONY: build

gen:
	go run cmd/static-gen/main.go --stderrthreshold=INFO --dir=posts/ --out=public_html/ --template=template.html
.PHONY: gen

run:
	 go run cmd/blog-server/main.go --root=public_html/ --stderrthreshold=INFO
.PHONY: run
