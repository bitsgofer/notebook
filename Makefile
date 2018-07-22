_root_dir=github.com/exklamationmark/notebook

PKG=./...
RUN=
ifneq ("$(RUN)","")
	_go_test_run_flag=-run $(RUN)
endif
GLOG=
test:
	go test -cover -v -race $(_root_dir)/$(PKG) ${_go_test_run_flag} $(GLOG)
	go vet $(_root_dir)/$(PKG)
.PHONY: test

gen:
	go run cmd/static-gen/main.go --stderrthreshold=INFO --dir=posts/ --out=public_html/ --template=template.html
.PHONY: gen

run:
	 go run cmd/blog-server/main.go --root=public_html/ --stderrthreshold=INFO
.PHONY: run
