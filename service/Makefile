#!/bin/sh

include ./local-run.sh

OSNAME=$(shell go env GOOS)
ARCH = $(shell go env GOARCH)

CMD_DIR=./cmd/service/.
ARTIFACT_DIR=./bin
ARTIFACT=${ARTIFACT_DIR}/svc
REPO=genvmoroz
IMAGE_NAME=lale-service
DOCKER_TAG=development

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

# The Linux build uses UPX to reduce the final binary size by ~70%
.PHONY: build
build:
	GOOS=${OSNAME} \
		go build \
		-o ${ARTIFACT} \
		${CMD_DIR}

.PHONY: build-docker
build-docker:
	docker build -t "${REPO}/${IMAGE_NAME}:${DOCKER_TAG}" .

.PHONY: push-docker
push-docker:
	docker push "${REPO}/${IMAGE_NAME}:${DOCKER_TAG}"

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

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	golangci-lint run --allow-parallel-runners -c ./.golangci.yml

.PHONY: gci
gci:
	go install github.com/luw2007/gci@latest
	gci write . --skip-generated
