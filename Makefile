.PHONY: help build test lint fmt clean docker-build docker-up docker-down coverage
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=wg-agent
VERSION?=v1.0.0
IMAGE_NAME?=wg-agent
IMAGE_TAG?=latest
DOCKER_REGISTRY?=ghcr.io/afrisinc

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -ldflags="-w -s -X main.Version=$(VERSION)" .
	@echo "✓ Build complete: $(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✓ Tests passed"

test-coverage: ## Run tests and generate coverage report
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...
	@echo "✓ Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet complete"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean
	@echo "✓ Clean complete"

tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	go mod tidy
	@echo "✓ Modules tidied"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "✓ Image built: $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "✓ Image pushed"

docker-up: ## Start Docker Compose stack
	@echo "Starting Docker Compose stack..."
	docker-compose up -d
	@echo "✓ Stack started. Check with: docker-compose ps"

docker-down: ## Stop Docker Compose stack
	@echo "Stopping Docker Compose stack..."
	docker-compose down
	@echo "✓ Stack stopped"

docker-logs: ## View Docker logs
	docker-compose logs -f wg-agent

security-scan: ## Run security scans
	@echo "Running security scans..."
	golangci-lint run ./...
	staticcheck ./...
	@echo "✓ Security scan complete"

install: build ## Install binary to $$GOPATH/bin
	@echo "Installing..."
	go install -ldflags="-w -s -X main.Version=$(VERSION)" .
	@echo "✓ Installed to $$GOPATH/bin"

run: ## Run the agent locally
	@echo "Running agent..."
	API_KEY="test-key" go run main.go

run-docker: docker-build ## Run agent in Docker
	@echo "Running in Docker..."
	docker run --rm -it \
		--cap-add NET_ADMIN \
		--cap-add SYS_MODULE \
		-p 9999:9999 \
		-e API_KEY="test-key" \
		-e WG_INTERFACE=wg0 \
		-v /etc/wireguard:/etc/wireguard:ro \
		$(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

release: ## Create a release build
	@echo "Creating release build..."
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 -ldflags="-w -s -X main.Version=$(VERSION)" .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 -ldflags="-w -s -X main.Version=$(VERSION)" .
	cd dist && sha256sum * > SHA256SUMS && cd ..
	@echo "✓ Release build complete in dist/"

all: fmt vet lint test build ## Run all checks and build

.PHONY: all build test lint fmt clean docker-build docker-push help
