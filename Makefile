.PHONY: build test lint clean validate-rules build-web release-build release

VERSION ?= $(shell cat VERSION 2>/dev/null || echo dev)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
WEB_LDFLAGS := -X 'main.buildVersion=$(VERSION)' -X 'main.buildTime=$(BUILD_TIME)'
RELEASE_OS ?= linux
RELEASE_ARCH ?= amd64

build:
	go build -o bin/metrics-analyzer ./cmd/metrics-analyzer

build-web:
	go build -ldflags "$(WEB_LDFLAGS)" -o bin/web-server ./web/server

release-build:
	mkdir -p dist
	GOOS=$(RELEASE_OS) GOARCH=$(RELEASE_ARCH) go build -o dist/sma-$(VERSION)-$(RELEASE_OS)-$(RELEASE_ARCH) ./cmd/metrics-analyzer
	GOOS=$(RELEASE_OS) GOARCH=$(RELEASE_ARCH) go build -ldflags "$(WEB_LDFLAGS)" -o dist/web-sma-$(VERSION)-$(RELEASE_OS)-$(RELEASE_ARCH) ./web/server

release:
	rm -rf dist
	mkdir -p dist
	RELEASE_OS=linux RELEASE_ARCH=amd64 $(MAKE) release-build
	RELEASE_OS=darwin RELEASE_ARCH=arm64 $(MAKE) release-build

	# calculate checksums
	sha256sum dist/* > dist/checksums.txt

test:
	go test -v ./...

lint:
	golangci-lint run || true

validate-rules:
	./bin/metrics-analyzer validate ./automated-rules

clean:
	rm -rf bin/

