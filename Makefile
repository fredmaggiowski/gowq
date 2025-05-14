all: test

.PHONY: test
test:
	go test ./... -race -cover -count=1

.PHONY: build
build:
	@go build -v ./...

assert-version:
	@test -n "$(VERSION)" || (echo "VERSION is required" ; exit 1)

.PHONY: version
version: assert-version
	@git tag v${VERSION}
