SHELL:=/bin/bash

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

lint: install-tools
	@echo "Linting go-client-langfuse module..."; \
	(go mod tidy && golangci-lint run --config .github/.golangci.yml --timeout 5m ./...);

.PHONY: test
test:
	@echo "Testing go-client-langfuse module..."; \
	(go mod download && go test -cover -coverprofile=coverage.out ./... && go tool cover -func coverage.out );

.PHONY: build
build:
	@echo "Building go-client-langfuse module..."; \
	(go build -o bin/go-client-langfuse-$(GOOS)-$(GOARCH) ./... );

.PHONY: tidy
tidy:
	@echo "Tidying go-client-langfuse module..."; \
	(go mod tidy);

.PHONY: fmt
fmt:
	@echo "Formatting go-client-langfuse module..."; \
	(gofmt -l -w -s . );

.PHONY: bump
bump:
	@echo "Bumping go-client-langfuse module..."; \
	(go get -u ./... && go mod tidy);

.PHONY: install-tools
install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b `go env GOPATH`/bin v2.5.0
