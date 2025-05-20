#!/bin/sh

SERVICE_MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
include $(SERVICE_MAKEFILE_DIR)local-run.sh $(SERVICE_MAKEFILE_DIR)../Makefile

OSNAME=$(shell go env GOOS)
ARCH = $(shell go env GOARCH)

CMD_DIR=$(SERVICE_MAKEFILE_DIR)cmd/service/.
ARTIFACT_DIR=$(SERVICE_MAKEFILE_DIR)bin
ARTIFACT=${ARTIFACT_DIR}/svc
DOCKER_REGISTRY=genvmoroz
SERVICE_TAG=service:development

.PHONY: ci
ci: lint test vulnerabilities_lookup

.PHONY: deps
deps:
	go mod tidy -v
	go mod verify

.PHONY: test
test:
	go test -v -cover ./... -count=1

.PHONY: bench
bench:
	go test ./... -bench=. -benchmem -run=^#

.PHONY: build
build:
	GOOS=${OSNAME} \
		go build \
		-o ${ARTIFACT} \
		${CMD_DIR}

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux \
	GOARCH=amd64 \
		go build \
		-o ${ARTIFACT}-linux-amd64 \
		${CMD_DIR}

.PHONY: build-docker
build-docker:
	docker build -t \
		"${DOCKER_REGISTRY}/${SERVICE_TAG}" \
		--build-arg ARTIFACT=$(ARTIFACT) \
		.

.PHONY: push-docker
push-docker:
	docker push "${DOCKER_REGISTRY}/${SERVICE_TAG}"

.PHONY: test_coverage
test_coverage:
	go test -coverprofile coverage.html ./... ; go tool cover -html=coverage.html
	rm ./coverage.html

.PHONY: gen
gen: protoc-gen
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go generate ./...

.PHONY: protoc-gen
protoc-gen:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/**.proto

vulnerabilities_lookup:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -test ./...

.PHONY: gci
gci:
	go install github.com/luw2007/gci@latest
	gci write . --skip-generated
