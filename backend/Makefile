ENTRYPOINT := ./cmd/
CONFIG := ./config/gohole.yaml
COMPOSE_FILE := ../compose.yaml

run: vet
	go run $(ENTRYPOINT) $(CONFIG)
.PHONY: run

setup:
	go get ./...
.PHONY: setup

update:
	go mod tidy
	go get -u ./...
.PHONY: update

vet:
	go fmt ./...
	go fix ./...
	go vet ./...
.PHONY: vet

bench:
	go test -bench=. ./...
.PHONY: bench

dockerup:
	docker compose -f $(COMPOSE_FILE) up -d
.PHONY: dockerup

dockerdown:
	docker compose -f $(COMPOSE_FILE) down
.PHONY: dockerdown
