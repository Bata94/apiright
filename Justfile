set dotenv-load := true

BINARY_NAME := "ApiRight"
DOCKER_REGISTRY := "ghcr.io/bata94/"
# EXPORT_RESULT := false # for CI please set EXPORT_RESULT to true


build-docker:
	docker build --target prod --tag $(BINARY_NAME) .

release-docker:
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest

mod-tidy:
	@echo "go mod tidy ..."
	go mod tidy

# Build the application
build-example:
	@echo "Building..."
	go build -o bin/main example/test/main.go

full-build-example:
	@echo "Full-Building..."
	CGO_ENABLED=0 go build -installsuffix 'static' -o bin/mainDocker main.go

# Run the application
run-example:
	go run example/test/main.go

# Test the application
test:
	@echo "Testing..."
	go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	rm -rf bin/*
	rm -rf tmp/*

# Live Reload
watch-example: build-example
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/air-verse/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

