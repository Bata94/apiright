set dotenv-load := true

BINARY_NAME := "ApiRight"

default:
  @just --list

mod-tidy:
	@echo "go mod tidy ..."
	go mod tidy

build-cli:
  @echo "Building CLI..."
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/apiright-cli cmd/apiright-cli/main.go

dev-templ:
  templ generate --watch --proxy="http://localhost:5500" --open-browser=false

dev-tailwind:
  tailwindcss -i ./example/ui/css/input.css -o ./example/assets/styles_gen.css --minify --watch

dev-air:
  ENV=dev air

# Develop with Example Application
[parallel]
dev: dev-templ dev-tailwind dev-air
  echo "Done..."

# Generate the UI routes
gen-example:
  @echo "Generating..."
  go run cmd/apiright-cli/main.go --input ./example/ui/pages -output ./example/ui-router/gen/routes_gen.go -package gen

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
	go test ./apiright.go -v
	go test ./pkg/... -v

# Clean the binary
clean:
  @echo "Cleaning..."
  rm -rf example/tmp/*
  rm -rf example/docs
  rm -rf bin/*
  rm -rf tmp/*
  rm -rf docs

check:
  go vet ./apiright.go
  go vet ./pkg/...

fmt:
  go fmt ./apiright.go
  go fmt ./pkg/...

simplify-fmt:
  gofmt -w -s ./apiright.go
  gofmt -w -s ./pkg/...

lint:
  golangci-lint run ./apiright.go
  golangci-lint run ./pkg/...

pre-release: fmt check lint test
  @echo "Ran check, fmt and lint"

vegeta method url max-workers duration:
  echo "{{method}} {{url}}" | vegeta attack -duration {{duration}} -rate 0 -max-workers {{max-workers}} | vegeta encode > results.json
  vegeta report results.json
