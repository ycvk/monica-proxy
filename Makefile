.PHONY: build build-all docker-build clean

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BIN_NAME = monica
BUILD_DIR = build

build:
	@rm -rf $(BUILD_DIR) || true
	@mkdir -p $(BUILD_DIR) || true
	@go mod tidy
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BIN_NAME) .

build-all:
	@$(MAKE) build GOOS=linux GOARCH=amd64
	@$(MAKE) build GOOS=linux GOARCH=arm64
	@$(MAKE) build GOOS=darwin GOARCH=arm64

docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 -t yourrepo/monica-proxy:latest --push .

clean:
	rm -rf $(BUILD_DIR)