# Common variables
VERSION := 0.0.1
BUILD_INFO := Manual build 
SRC_DIR := ./cmd

.PHONY: help build run lint lint-fix
.DEFAULT_GOAL := help

help:  ## This help message :)
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint:  ## Lint & format, will not fix but sets exit code on error
	@which golangci-lint > /dev/null || go get github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run $(SRC_DIR)/...

lint-fix:  ## Lint & format, will try to fix errors and modify code
	@which golangci-lint > /dev/null || go get github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run $(SRC_DIR)/... --fix

build:  ## Run a local build without a container
	go build -o bin/runr $(SRC_DIR)/...

run:  ## Run application, used for local development
	air -c .air.toml