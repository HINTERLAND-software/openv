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
	# Create a temporary directory for generating docs
	mkdir -p tmp_docs
	# Generate the documentation in the temporary directory
	go run main.go gen-doc -o tmp_docs
	# If the generation is successful, replace the existing docs directory
	if [ $$? -eq 0 ]; then \
		rm -rf docs; \
		mv tmp_docs docs; \
	else \
		rm -rf tmp_docs; \
		echo "Documentation generation failed. Existing docs are preserved."; \
	fi
