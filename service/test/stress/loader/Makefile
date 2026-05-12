#!/bin/sh

.PHONY: deps
deps:
	go mod tidy
	go mod verify

.PHONY: gen
gen:
	go install golang.org/x/tools/cmd/stringer@latest
	go generate ./...

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --allow-parallel-runners -c ./../../../.golangci.yml

.PHONY: gci
gci:
	go install github.com/luw2007/gci@latest
	gci write . --skip-generated
