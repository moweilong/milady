SHELL := /bin/bash

PROJECT_NAME := "github.com/moweilong/milady"
PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/ | grep -v /api/ | grep -v /cmd/)

# delete the templates code start
.PHONY: install
# Installation of dependent plugins
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/srikrsna/protoc-gen-gotag@latest
	go install github.com/go-dev-frame/sponge/cmd/protoc-gen-go-gin@latest
	go install github.com/go-dev-frame/sponge/cmd/protoc-gen-go-rpc-tmpl@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@v1.8.12
# 	go install github.com/ofabry/go-callvis@latest
	go install golang.org/x/pkgsite/cmd/pkgsite@latest
# delete the templates code end

.PHONY: ci-lint
# Check code formatting, naming conventions, security, maintainability, etc. the rules in the .golangci.yml file
ci-lint:
	@gofmt -s -w .
	golangci-lint run ./...

.PHONY: build
# Build serverNameExample_mixExample for linux amd64 binary
build:
	@echo "building 'serverNameExample_mixExample', linux binary file will output to 'cmd/serverNameExample_mixExample'"
	@cd cmd/serverNameExample_mixExample && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# delete the templates code start

.PHONY: build-milady
# Build milady for linux amd64 binary
build-milady:
	@echo "building 'milady', linux binary file will output to 'cmd/milady'"
	@cd cmd/milady && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "all=-s -w"
