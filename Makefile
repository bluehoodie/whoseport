.PHONY: build install clean test

# Build binary for local testing
build:
	go build -o whoseport .

# Install binary to $GOBIN (or $GOPATH/bin or $HOME/go/bin)
install:
	go install .

# Remove locally built binary
clean:
	rm -f whoseport

# Run tests
test:
	go test ./...
