LOCAL_BIN:=$(CURDIR)/bin
PATH:=$(LOCAL_BIN):$(PATH)
ENTRYPOINT := ./cmd/
CONFIG := ./config/gohole.yaml
COMPOSE_FILE := ../compose.yaml
UNIT_TEST_DIR := ./...
OUTPUT_DIR := ./output
GOLANGCI_LINT_VERSION=v2.9.0

default: run

help: ### Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

setup: download-lint ### Install dependencies
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go mod download
.PHONY: setup

download-lint: ### Download golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_LINT_VERSION)
.PHONY: download-lint

update: ### Update existing dependencies and install new ones
	GOBIN=$(LOCAL_BIN) go mod tidy
	GOBIN=$(LOCAL_BIN) go get -u ./...
.PHONY: update

run: format ### Run the application
	go run $(ENTRYPOINT) $(CONFIG)
.PHONY: run

build: ### Build the application
	GOBIN=$(LOCAL_BIN) go build -o ${OUTPUT_DIR}/gohole ${ENTRYPOINT}
.PHONY: build

generate: ### Generate code
	go generate ./...
.PHONY: generate

test: generate ### Run unit tests
	go tool gotest.tools/gotestsum --format testdox -- -v -cover -race ${UNIT_TEST_DIR}
.PHONY: test

coverage: generate ### Generate test coverage html report
	@mkdir -p ${OUTPUT_DIR}
	go tool gotest.tools/gotestsum --format github-actions -- -v -coverprofile=${OUTPUT_DIR}/coverage.out -cover -race ${UNIT_TEST_DIR}
	go tool cover -html=${OUTPUT_DIR}/coverage.out
.PHONY: coverage

bench: ### Run benchmarks
	go test -bench=. ./...
.PHONY: bench

format:
	@printf "Running go fix... "
	@go fix ./...
	@printf "done.\n"
	@printf "Formatting code... "
	@go tool golines -w .
	@go fmt ./...
	@go tool goimports -w .
	@printf "done.\n"
	@printf "Running go vet... "
	@go vet ./...
	@printf "done.\n"
.PHONY: format

lint: ### Linter check
	GOBIN=$(LOCAL_BIN) golangci-lint run
.PHONY: lint

dockerup: ### Start the application using Docker Compose
	docker compose -f $(COMPOSE_FILE) up -d
.PHONY: dockerup

dockerdown: ### Stop the application using Docker Compose
	docker compose -f $(COMPOSE_FILE) down
.PHONY: dockerdown

clean: ### Clean the project
	go clean
	rm -rf ${OUTPUT}
	rm -rf ${LOCAL_BIN}
.PHONY: clean