all: test

.PHONY: test
test:
	go test ./... -race -cover -count=1

.PHONY: build
build:
	@go build -v ./...
