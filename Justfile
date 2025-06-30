set dotenv-load := true

BINARY_NAME := "ApiRight"

default:
  @just --list

mod-tidy:
	@echo "go mod tidy ..."
	go mod tidy

# Develop with Example Application
dev:
  ENV=dev air

# Build the application
build-example:
	@echo "Building..."
	go build -o bin/example-main example/main.go

# Run the application
run-example:
	go run example/main.go

# Test the example application
test-example:
  example/curl_test.sh

# Test the library
test:
	@echo "Testing..."
	go test ./... -v

# Clean the binary
clean:
  @echo "Cleaning..."
  rm -rf example/tmp/*
  rm -rf example/docs
  rm -rf bin/*
  rm -rf tmp/*
  rm -rf docs

check:
  go vet ./...

fmt:
  go fmt ./...

simplify-fmt:
  gofmt -w -s .

lint:
  golangci-lint run ./...

pre-release: clean mod-tidy fmt check lint test
  @echo "Ran check, fmt and lint"
