#!/bin/sh

include ./scripts/linting.mk
include ./scripts/gen.mk

OSNAME=$(shell go env GOOS)
ARCH = $(shell go env GOARCH)

CMD_DIR=./cmd/service/.
ARTIFACT_DIR=./bin
ARTIFACT=${ARTIFACT_DIR}/svc
REPO=genmoroz
IMAGE_NAME=lale-tg-client
DOCKER_TAG=development

.PHONY: ci
ci: lint test

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

.PHONY: run-infra
run-infra:
	docker compose -f ./deployments/docker-compose.yml up -d --build --remove-orphans --force-recreate

.PHONY: stop-infra
stop-infra:
	docker compose -f ./deployments/docker-compose.yml down

.PHONY: run-locally
run-locally:
	${ARTIFACT}
