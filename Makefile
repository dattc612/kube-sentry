BIN_NAME   = app
export GO111MODULE = on
IMAGE_NAME ?= KubeSentry
IMAGE_TAG  ?= latest

.DEFAULT_GOAL := help

.PHONY: all
all: init build ## Run full initialization and build both client and server

.PHONY: init
init: init-client init-server ## Install dependencies for both client and server

.PHONY: init-client
init-client: ## Install client dependencies using yarn
	@echo "> Installing the client dependencies ..."
	@yarn install -D --unsafe-perm

.PHONY: init-server
init-server: ## Install server dependencies and live-reload tool (fresh)
	@echo "> Installing the server dependencies ..."
	@go mod tidy -v
	@go get -v ./...
	@go install github.com/gravityblast/fresh@latest

.PHONY: build
build: build-client build-server ## Build both client and server binaries

.PHONY: build-client
build-client: ## Compile frontend client assets
	@echo "> Building the client ..."
	yarn run build

.PHONY: build-server
build-server: ## Compile backend Go binary into ./bin/app
	@echo "> Building the server binary ..."
	@rm -rf bin && go build -o bin/${BIN_NAME} .

.PHONY: run-dev
run-dev: ## Run frontend and backend concurrently with live reload
	@echo "> Running Frontend & Backend ..."
	./node_modules/.bin/concurrently "yarn run dev" "fresh"

.PHONY: docker-build
docker-build: ## Build the Docker container image
	@echo "==> Building Docker image: $(IMAGE_NAME):$(IMAGE_TAG)"
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

.PHONY: help
help: ## Display this help screen
	@echo ""
	@echo "Usage:"
	@echo "  make \033[36m<target>\033[0m"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""