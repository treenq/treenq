LAST_TAG := $(shell git describe --tags --abbrev=0)

COMMIT_SHA := $(shell git rev-parse --short HEAD)

CURRENT_BRANCH := $(shell git branch --show-current)

ifeq ($(CURRENT_BRANCH),main)
    VERSION_STRING := $(LAST_TAG)
else
    VERSION_STRING := develop-$(COMMIT_SHA)
endif

build:
	go build -ldflags="-X 'github.com/treenq/treenq/src/handlers.version=$(VERSION_STRING)'" ./cmd/server/main.go

