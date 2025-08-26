# Makefile for secret-storage-client

# Variables
PROTO_DIR = ./pkg/proto
PROTO_FILE = $(PROTO_DIR)/secretstorage.proto
GO_OUT_DIR = $(PROTO_DIR)

# Default target
.PHONY: all
all: proto build

# Generate protobuf code
.PHONY: proto
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)

# Build the application
.PHONY: build
build:
	@echo "Building application..."
	go build -o bin/secret-storage-client cmd/main.go

# Run the application
.PHONY: run
run: build
	./bin/secret-storage-client

# Clean generated files
.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	rm -f $(PROTO_DIR)/*.pb.go
	rm -f bin/secret-storage-client

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Install protobuf tools
.PHONY: install-tools
install-tools:
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Test
.PHONY: test
test:
	go test ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Generate protobuf code and build"
	@echo "  proto         - Generate protobuf code"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Clean generated files"
	@echo "  deps          - Install dependencies"
	@echo "  install-tools - Install protobuf tools"
	@echo "  test          - Run tests"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  help          - Show this help"