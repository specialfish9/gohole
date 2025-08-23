ENTRYPOINT := ./cmd/

run:
	go run $(ENTRYPOINT)

setup:
	go get ./...

update:
	go mod tidy
	go get -u ./...

