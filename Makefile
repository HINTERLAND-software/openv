VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null)
# DEFAULT VERSION TO 0.0.1 IF NO TAGS ARE AVAILABLE
ifeq ($(VERSION),)
	VERSION := 0.0.1
endif

COMMIT_HASH ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
PKG := github.com/hinterland-software/openv
LDFLAGS := -X '$(PKG)/internal/version.Version=$(VERSION)' \
           -X '$(PKG)/internal/version.CommitHash=$(COMMIT_HASH)' \
           -X '$(PKG)/internal/version.BuildTime=$(BUILD_TIME)'

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/openv main.go

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" 