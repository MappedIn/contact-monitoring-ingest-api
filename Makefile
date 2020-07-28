# https://github.com/azer/go-makefile-example/blob/master/Makefile

-include .env

PROJECTNAME := contact-monitoring-ingest-api

# Go related variables.
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := ./cmd/main.go

# Make is verbose in Linux. Make it silent.
# MAKEFLAGS += --silent

## install: installs go dependencies (but not go itself)
install:
	@go mod download

## run: runs from sourcecode; dev only
run:
	@go run $(GOFILES)

## build: builds binary for this project
build:
	@echo "  >  Building binary..."
	@go build \
		-a \
		-tags netgo,static_all \
		-o $(GOBIN)/server \
		$(GOFILES)

## start: runs the previously built binary
start:
	@$(GOBIN)/server

## docker-build: builds docker iamge for this project
docker-build:
	@docker build -t $(PROJECTNAME) .

## docker-run: runs previously built docker image for this project
docker-run:
	@docker run \
		-it \
		--rm \
		--env-file=./.env \
		-p $(PORT):80 \
		$(PROJECTNAME)

## test: runs all tests with verbose output
test:
	@go test ./... -v

## test-cover: runs all tests with verbose output and coverage
test-cover:
	@go test ./... -v -cover

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
