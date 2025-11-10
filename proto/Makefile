.PHONY: proto proto-lint proto-breaking clean install-buf

# Generate Go code from protobuf definitions
# Uses buf if available, falls back to protoc
proto:
	@echo "Generating Go protobuf code..."
	@if command -v buf >/dev/null 2>&1; then \
		buf generate; \
	else \
		echo "buf not found, using protoc..."; \
		protoc --go_out=generated --go_opt=paths=source_relative --proto_path=types types/plugin.proto; \
	fi
	@echo "Running go mod tidy..."
	@cd ../.. && go mod tidy
	@echo "Done!"

# Lint protobuf files (requires buf)
proto-lint:
	@if command -v buf >/dev/null 2>&1; then \
		buf lint; \
	else \
		echo "Error: buf is required for linting. Install with: brew install bufbuild/buf/buf"; \
		exit 1; \
	fi

# Check for breaking changes (requires buf)
proto-breaking:
	@if command -v buf >/dev/null 2>&1; then \
		buf breaking --against '.git#branch=main'; \
	else \
		echo "Error: buf is required for breaking change detection. Install with: brew install bufbuild/buf/buf"; \
		exit 1; \
	fi

# Install buf CLI
install-buf:
	@echo "Installing buf..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install bufbuild/buf/buf; \
	else \
		echo "Homebrew not found. Installing via go install..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	@echo "buf installed successfully!"

# Clean generated files
clean:
	@rm -rf generated/*.pb.go
	@echo "Cleaned generated files"

# Full workflow: lint + generate
all: proto-lint proto

# Help message
help:
	@echo "Available targets:"
	@echo "  make proto           - Generate Go code from .proto files"
	@echo "  make proto-lint      - Lint protobuf files (requires buf)"
	@echo "  make proto-breaking  - Check for breaking changes (requires buf)"
	@echo "  make install-buf     - Install buf CLI"
	@echo "  make clean           - Remove generated files"
	@echo "  make all             - Lint and generate"

