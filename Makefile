ENTRYPOINT := ./cmd/

run:
	go run $(ENTRYPOINT)

update:
	go mod tidy
	go get -u ./...

