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
