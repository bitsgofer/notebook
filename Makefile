PKG=./...
RUN=
ifneq ("$(RUN)","")
	RUN=-run $(RUN)
endif
test:
	go test -v -race $(PKG) $(RUN)
.PHONY: test
