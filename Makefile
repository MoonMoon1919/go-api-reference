# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVERSION=`cat go.mod | grep 'go\s\d.' | cut -d ' ' -f2`

GOENVCMD=goenv
BUILDSHA=`git rev-parse --short HEAD`

BINARY_NAME=`pwd | xargs -L1 basename`
ADMIN_BINARY_NAME=$(BINARY_NAME)-admin
ADD_USER_BINARY_NAME="add-user"
DELETE_USER_BINARY_NAME="delete-user"

# Check if required tools are installed
.PHONE: check-goenv
check-goenv:
	@which $(GOENVCMD) >/dev/null 2>&1 || \
		(echo "ERROR: goenv is not installed or not in PATH" && exit 1)

.PHONY: check-tools
check-tools:
	@which $(GOCMD) >/dev/null 2>&1 || \
		(echo "ERROR: Go is not installed or not in PATH" && exit 1)

# Format all go files
.PHONY: fmt
fmt: check-tools
	@$(GOFMT) ./...

# Run go vet
.PHONY: vet
vet: check-tools
	@$(GOVET) ./...

# Download dependencies
.PHONY: deps
deps: check-tools
	@$(GOMOD) download
	@$(GOMOD) verify

# Build the binary
.PHONY: build/api
build/api: check-tools fmt vet
	@$(GOBUILD) -ldflags "-X 'github.com/moonmoon1919/go-api-reference/internal/build.VERSION=$(BUILDSHA)'" -o $(BINARY_NAME) cmd/api/main.go

.PHONY: build/admin-api
build/admin-api: check-tools fmt vet
	@$(GOBUILD) -ldflags "-X 'github.com/moonmoon1919/go-api-reference/internal/build.VERSION=$(BUILDSHA)'" -o $(ADMIN_BINARY_NAME) cmd/admin_api/main.go

.PHONY: build/add-user
build/add-user: check-tools fmt vet
	@$(GOBUILD) -ldflags "-X 'github.com/moonmoon1919/go-api-reference/internal/build.VERSION=$(BUILDSHA)'" -o $(ADD_USER_BINARY_NAME) cmd/add_user/main.go

.PHONY: build/delete-user
build/delete-user: check-tools fmt vet
	@$(GOBUILD) -ldflags "-X 'github.com/moonmoon1919/go-api-reference/internal/build.VERSION=$(BUILDSHA)'" -o $(DELETE_USER_BINARY_NAME) cmd/delete_user/main.go

# Run the application
.PHONY: run
run: check-tools
	@$(GOCMD) run ./cmd/api/main.go

# Clean build artifacts
.PHONY: clean
clean: check-tools
	@rm -f $(BINARY_NAME)
	@rm -f $(DELETE_USER_BINARY_NAME)
	@rm -f $(ADD_USER_BINARY_NAME)
	@go clean

# Run tests
.PHONY: test/unit
test/unit: check-tools
	@$(GOTEST) -v ./...


# Run unit tests with coverage
.PHONY: test/coverage
test/coverage: check-tools
	@$(GOTEST) -v -cover ./...

# Run integration tests
.PHONY: test/integration
test/integration: check-tools
	@TEST_TYPE=INTEGRATION $(GOTEST) -v ./... -run TestIntegration

.PHONY: init-shell
init-shell: check-goenv
	@$(GOENVCMD) local $(GOVERSION)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build/api         - Builds the api"
	@echo "  build/admin-api   - Builds the admin-api"
	@echo "  build/add-user    - Builds the event listener for adding users"
	@echo "  build/delete-user - Builds the event listener for deleting users"
	@echo "  clean             - Removes build artifacts"
	@echo "  deps              - Downloads and verify dependencies"
	@echo "  fmt               - Formats Go source files"
	@echo "  help              - Shows this help message"
	@echo "  run               - Runs the application"
	@echo "  test/unit         - Runs unit tests"
	@echo "  test/coverage     - Runs unit tests with coverage"
	@echo "  test/integration  - Runs integration tests"
	@echo "  vet               - Runs go vet"
	@echo "  init-shell        - Sets goversion using goenv"

# Default target
.DEFAULT_GOAL := help
