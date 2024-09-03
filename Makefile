.PHONY: lint

DB_DSN ?= "postgres://postgres@localhost:5432/tq?sslmode=disable"

LAST_TAG := $(shell git describe --tags --abbrev=0)

COMMIT_SHA := $(shell git rev-parse --short HEAD)

CURRENT_BRANCH := $(shell git branch --show-current)

ifeq ($(CURRENT_BRANCH),main)
    VERSION := $(LAST_TAG)
else
    VERSION := develop-$(COMMIT_SHA)
endif

build:
	go build -ldflags="-X 'github.com/treenq/treenq/src/handlers.version=$(VERSION)'" ./cmd/server/main.go


install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	go install golang.org/x/tools/cmd/goimports@v0.22.0

lint:
	@if goimports -l . | grep -q '.'; then \
		echo "goimports failed"; \
		exit 1; \
	fi
	golangci-lint run

migrate_new:
	migrate create -ext sql -dir migrations -seq -digits 4 ${MNAME}

migrate_up:
	migrate -path migrations -database ${DB_DSN} up

migrate_down:
	migrate -path migrations -database ${DB_DSN} down

migrate_fix:
	migrate -path migrations -database ${DB_DSN} force ${V}

migrate_v:
	migrate -path migrations -database ${DB_DSN} version
