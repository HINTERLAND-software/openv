VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null)
# DEFAULT VERSION TO dev IF NO TAGS ARE AVAILABLE
ifeq ($(VERSION),)
	VERSION := dev
endif

COMMIT_HASH ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -X 'github.com/hinterland-software/openv/internal/version.Version=$(VERSION)' \
           -X 'github.com/hinterland-software/openv/internal/version.CommitHash=$(COMMIT_HASH)' \
           -X 'github.com/hinterland-software/openv/internal/version.BuildTime=$(BUILD_TIME)'

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/openv main.go

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)"

.PHONY: lint
lint:
	golangci-lint run

.PHONY: format
format:
	go fmt ./...

.PHONY: docs
docs:
	go run main.go gen-doc
