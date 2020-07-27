IMPORT_PATH=github.com/bitsgofer/notebook
define GO_ENV
	GOPROXY=https://proxy.golang.org,direct \
	GOPRIVATE=github.com/bitsgofer/*
endef
GO=${GO_ENV}; go
define GORELEASER
	@echo "Using Github token: $$GITHUB_TOKEN";
	docker run -v ${PWD}:/go/src/${IMPORT_PATH} \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/${IMPORT_PATH} \
		-e GITHUB_TOKEN \
		-e DOCKER_USERNAME \
		-e DOCKER_PASSWORD \
		-e DOCKER_REGISTRY \
		goreleaser/goreleaser:v0.139
endef

.mk_state:
	mkdir -p .mk_state

# Empty target to record events: http://www.gnu.org/software/make/manual/make.html#Empty-Targets
.mk_state/goreleaser.docker: .mk_state
	docker pull goreleaser/goreleaser:v0.139
	touch .mk_state/goreleaser.docker

# Phone targets: http://www.gnu.org/software/make/manual/make.html#Phony-Targets
.PHONY: clean
clean:
	rm -rf .mk_state


# $(GORELEASER) --snapshot --skip-publish --rm-dist

.PHONY: test
PKG=./...
RUN=
ifneq ("$(RUN)","")
	_test_regex_flag=-run $(RUN)
endif
test:
	$(GO) test -cover -v -race ${IMPORT_PATH}/${PKG} ${_test_regex_flag}
.PHONY: test


.PHONY: vet
vet:
	$(GO) vet ${IMPORT_PATH}/...


.PHONY: build
build:
	$(GO) build ${IMPORT_PATH}/${PKG}


# PKG=./...
# RUN=
# ifneq ("$(RUN)","")
# 	_go_test_run_flag=-run $(RUN)
# endif
# GLOG=
#
# all: test build gen
# .PHONY: all
#
# test:
# 	go test -cover -v -race $(_pkg)/$(PKG) ${_go_test_run_flag} $(GLOG)
# 	go vet $(_pkg)/$(PKG)
# .PHONY: test
#
# build:
# 	mkdir -p build
# 	go build -o build/notebook ${_pkg}/cmd/notebook
# .PHONY: build
#
# gen.assets:
# 	minify assets/*.css > public_html/builtin.css
# 	minify assets/*.js > public_html/builtin.js
# 	cp assets/favicon.ico public_html/favicon.ico
# .PHONY: gen.assets
#
# gen: clean gen.assets
# 	./build/notebook generate --post.dir=posts/ --html.dir=public_html/ --post.template=template.html
# .PHONY: gen
#
# serve.local:
# 	sudo ./build/notebook serve --html.dir=public_html/
# .PHONY: serve.local
#
# run: build gen serve.local
# .PHONY: run
#
# clean:
# 	rm -rf public_html/*/
# 	rm -f public_html/*.css
# 	rm -f public_html/*.js
# 	rm -f public_html/*.ico
# .PHONY: clean
