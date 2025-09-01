.PHONY: test build clean lint fmt check

test:
	go test -v ./...

build:
	go build -v ./...

clean:
	go clean ./...

lint:
	go vet ./...

fmt:
	treefmt --allow-missing-formatter

check: fmt lint test
