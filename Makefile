.PHONY: test clean binaries

CIRCLE_BUILD_NUM ?= 0
VERSION = 0.0.$(CIRCLE_BUILD_NUM)-$(shell git rev-parse --short HEAD)

GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

GO_LDFLAGS := " \
    -X main.Version=$(VERSION) \
    -X main.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) \
"

GO_SOURCES=$(shell find . -name '*.go')

default: test

clean:
	rm -rf pkg

test:
	go test -tags "$(GO_TAGS)" ./...

pkg/bazel-remote-proxy-Darwin_x86_64: $(GO_SOURCES)
		@echo "==> Building $@ with tags $(GO_TAGS)..."
		@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
				go build \
				-ldflags $(GO_LDFLAGS) \
				-tags "$(GO_TAGS)" \
				-o "$@"

pkg/bazel-remote-proxy-Linux_x86_64: $(GO_SOURCES)
		@echo "==> Building $@ with tags $(GO_TAGS)..."
		@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
				go build \
				-ldflags $(GO_LDFLAGS) \
				-tags "$(GO_TAGS)" \
				-o "$@"

binaries: pkg/bazel-remote-proxy-Linux_x86_64 pkg/bazel-remote-proxy-Darwin_x86_64
