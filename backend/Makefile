ENTRYPOINT := ./cmd/
CONFIG := ./config/gohole.yaml
COMPOSE_FILE := ../compose.yaml
UNIT_TEST_DIR := ./...
OUTPUT_DIR := ./output

default: run

setup:
	go mod download
.PHONY: setup

update:
	go mod tidy
	go get -u ./...
.PHONY: update

run: vet
	go run $(ENTRYPOINT) $(CONFIG)
.PHONY: run

build:
	go build -o ${OUTPUT_DIR}/gohole ${ENTRYPOINT}
vet:
	go fmt ./...
	go fix ./...
	go vet ./...
.PHONY: vet

generate:
	go generate ./...
.PHONY: generate

test: generate
	go tool gotest.tools/gotestsum --format testdox -- -v -cover -race ${UNIT_TEST_DIR}
.PHONY: test

coverage: generate ### Generate test coverage html report
	@mkdir -p ${OUTPUT_DIR}
	go tool gotest.tools/gotestsum --format github-actions -- -v -coverprofile=${OUTPUT_DIR}/coverage.out -cover -race ${UNIT_TEST_DIR}
	go tool cover -html=${OUTPUT_DIR}/coverage.out
.PHONY: coverage

bench:
	go test -bench=. ./...
.PHONY: bench

dockerup:
	docker compose -f $(COMPOSE_FILE) up -d
.PHONY: dockerup

dockerdown:
	docker compose -f $(COMPOSE_FILE) down
.PHONY: dockerdown
