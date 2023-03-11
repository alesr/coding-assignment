.DEFAULT_GOAL := help

NAME = $(shell basename $(PWD))

.PHONY: help
help:
	@echo "------------------------------------------------------------------------"
	@echo "${NAME}"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[a-zA-Z0-9_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the application
	@GOOS=linux go build -o $(NAME) main.go

.PHONY: run
run: build ## Run the application on a Docker container (requires Docker)
	@docker-compose -f build/docker-compose.yml up go-alessandro-resta --force-recreate --build

.PHONY: lint
lint: ## Run go fmt and go vet
	@go fmt ./...
	@go vet ./...

.PHONY: test-unit
test-unit: ## Run unit tests
	@go test -v -race -vet=all -count=1 -timeout 60s ./...

.PHONY: test ## Run static analysis and unit tests
test: lint test-unit
