ENTRYPOINT := ./cmd/
CONFIG := ./config/gohole.conf
COMPOSE_FILE := ../compose.yaml

run:
	go run $(ENTRYPOINT) $(CONFIG)
.PHONY: run

setup:
	go get ./...
.PHONY: setup

update:
	go mod tidy
	go get -u ./...
.PHONY: update

dockerup:
	docker compose -f $(COMPOSE_FILE) up -d
.PHONY: dockerup

dockerdown:
	docker compose -f $(COMPOSE_FILE) down
.PHONY: dockerdown
