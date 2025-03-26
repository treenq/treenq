DB_DSN ?= "postgres://postgres@localhost:5432/tq?sslmode=disable"

COMPOSE_CMD := $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; else echo "docker-compose"; fi)

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
	$(COMPOSE_CMD) -p treenq -f docker-compose.yaml -f docker-compose.e2e.yaml up kube -d 
	sleep 1
	sed -i  's#https://127.0.0.1:6443#https://kube:6443#g' k3s_data/k3s/k3s.yaml
	$(COMPOSE_CMD) -p treenq -f docker-compose.yaml -f docker-compose.e2e.yaml up -d --build
	while [ -z '$$(docker ps -q --filter "name=treenq-server")' ]; do sleep 1; done	
	docker cp k3s_data/k3s/k3s.yaml $$(docker ps -q --filter "name=treenq-server"):/app/kubeconfig.yaml
	$(COMPOSE_CMD) -p treenq -f docker-compose.yaml -f docker-compose.e2e.yaml restart server 
	@echo "Checking e2e test environment is running..."
	until $$(curl --output /dev/null --silent --fail http://localhost:8000/healthz); do printf '.'; sleep 1; done && echo "Service Ready!"
	echo 'Service has been started'

stop-e2e-test-env:
	$(COMPOSE_CMD) -f docker-compose.yaml -f docker-compose.e2e.yaml down

run-e2e-tests: start-e2e-test-env
	go test -v -count=1 -race ./e2e/...
	make stop-e2e-test-env

unit-tests:
	go test $$(go list ./... | grep -v e2e)
