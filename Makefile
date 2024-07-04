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


install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	go install golang.org/x/tools/cmd/goimports@v0.22.0

lint:
	goimports -l -w . && golangci-lint run
