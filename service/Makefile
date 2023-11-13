#!/bin/sh

include ./scripts/linting.mk
include ./scripts/gen.mk

OSNAME=$(shell go env GOOS)
ARCH = $(shell go env GOARCH)

CMD_DIR=./cmd/service/.
ARTIFACT_DIR=./bin
ARTIFACT=${ARTIFACT_DIR}/svc
REPO=genmoroz
IMAGE_NAME=lale-service
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
	docker compose -f ./deployments/docker-compose.yml up -d --remove-orphans

.PHONY: stop-infra
stop-infra:
	docker compose -f ./deployments/docker-compose.yml down

.PHONY: run-locally
run-locally:
	APP_LOG_LEVEL=DEBUG \
	APP_GRPC_PORT=8080 \
	APP_HIPPO_SENTENCE_HOST="https://www.wordhippo.com" \
	APP_HIPPO_SENTENCE_RETRIES=3 \
	APP_HIPPO_SENTENCE_TIMEOUT=5s \
	APP_YOUR_DICTIONARY_SENTENCE_HOST="https://sentence.yourdictionary.com" \
	APP_YOUR_DICTIONARY_SENTENCE_RETRIES=3 \
	APP_YOUR_DICTIONARY_SENTENCE_TIMEOUT=5s \
	APP_MONGO_CARD_PROTOCOL="mongodb" \
	APP_MONGO_CARD_HOST="localhost" \
	APP_MONGO_CARD_PORT=27017 \
	APP_MONGO_CARD_URI_PARAMS="" \
	APP_MONGO_USER="admin" \
	APP_MONGO_PASS="pass" \
	APP_MONGO_CARD_DATABASE="dictionary" \
	APP_MONGO_CARD_COLLECTION="cards" \
	APP_REDIS_DB=0 \
	APP_REDIS_HOST="localhost" \
	APP_REDIS_INSECURE_SKIP_VERIFY=true \
	APP_REDIS_PASS="pass" \
	APP_REDIS_PORT=6379 \
	APP_REDIS_USE_TLS=false \
	APP_USER_SESSION_EXPIRATION=1h \
	APP_USER_SESSION_KEY_PREFIX="user-session" \
	APP_DICTIONARY_HOST="https://api.dictionaryapi.dev" \
	APP_GOOGLE_PROJECT_KEY_JSON=${APP_GOOGLE_PROJECT_KEY_JSON} \
	${ARTIFACT}

.PHONY: vulnerabilities_lookup
vulnerabilities_lookup:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -test ./...
