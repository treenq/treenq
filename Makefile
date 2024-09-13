DB_DSN ?= "postgres://postgres@localhost:5432/tq?sslmode=disable"

VERSION := $(shell \
	if [ "$(shell git branch --show-current)" = "main" ]; then \
		git describe --tags; \
	else \
		echo "$(shell git branch --show-current)-$(shell git rev-parse --short HEAD)"; \
	fi)

version:
	@echo $(VERSION)

build:
	go build -ldflags="-X 'github.com/treenq/treenq/src/handlers.version=$(VERSION)'" ./cmd/server/main.go

install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	go install golang.org/x/tools/cmd/goimports@v0.25.0
	go install github.com/go-delve/delve/cmd/dlv@v1.23.0

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

start-e2e-test-env:
	docker-compose -f docker-compose.e2e.yaml up -d --build
	@echo "Checking e2e test environment is running..."
	until $$(curl --output /dev/null --silent --fail http://localhost:8000/healthz); do printf '.'; sleep 1; done && echo "Service Ready!"
	echo 'Service has been started'

stop-e2e-test-env:
	docker-compose -f docker-compose.e2e.yaml down

run-e2e-tests: start-e2e-test-env
	go test -v -count=1 -race ./e2e/...
	make stop-e2e-test-env

unit-tests:
	go test $$(go list ./... | grep -v e2e)
