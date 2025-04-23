SHELL=/bin/bash
REWARD_BINARY=authApp
PROJECT_ROOT=..
BUILD_DIR=$(PROJECT_ROOT)/build

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_auth
	@echo "Stopping docker images (if running...)"
	docker compose down
	@echo "Building (when required) and starting docker images..."
	docker compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker compose down
	@echo "Done!"

## build_auth: builds the reward binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	@mkdir -p $(BUILD_DIR)
	cd $(PROJECT_ROOT) && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o $(BUILD_DIR)/$(REWARD_BINARY) ./cmd/app
	@echo "Done!"

## clean: remove built binary
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"