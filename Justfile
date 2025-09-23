set dotenv-load := true

default:
  @just --list

mod-tidy:
	@echo "go mod tidy ..."
	go mod tidy

# Build the CLI for the current platform
build-cli:
  @echo "Building CLI..."
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli ./apiright.go

# Build the CLI for all platforms
build-cli-all:
  @echo "Building CLI for all platforms..."
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli-linux-amd64 ./apiright.go
  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli-linux-arm64 ./apiright.go
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli-windows-amd64.exe ./apiright.go
  GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli-darwin-amd64 ./apiright.go
  GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli-darwin-arm64 ./apiright.go

dev-templ:
  templ generate --watch --proxy="http://localhost:5500" --open-browser=false

dev-tailwind:
  tailwindcss -i ./example/ui/css/input.css -o ./example/assets/styles_gen.css --minify --watch

dev-air:
  GOEXPERIMENT=greenteagc ENV=dev air

# Develop with Example Application
[parallel]
dev: dev-templ dev-tailwind dev-air
  echo "Done..."

# Generate the UI routes
gen-example:
  @echo "Generating..."
  go run ./apiright.go generate -i ./example/ui/pages -o ./example/uirouter/routes_gen.go

# Build the application
build-example: gen-example
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
  rm -rf example/uirouter
  rm -rf bin/*
  rm -rf tmp/*
  rm -rf docs

check:
  go vet ./...

fmt:
  go fmt ./...

simplify-fmt:
  gofmt -w -s ./...

lint:
  golangci-lint run ./...

pre-release: fmt check lint test
  @echo "Ran check, fmt and lint"

vegeta method url max-workers duration:
  echo "{{method}} {{url}}" | vegeta attack -duration {{duration}} -rate 0 -max-workers {{max-workers}} | vegeta encode > tmp/results.json
  vegeta report tmp/results.json
