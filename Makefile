DOCKER_REGISTRY=ghcr.io/genvmoroz
SERVICE_TAG=lale-service:latest
TG_CLIENT_TAG=lale-tg-client:latest

.PHONY: all update_deps build_and_push_latest update_deps_and_run_tests build_and_push

all: update_deps build_and_push_latest

update_deps:
	$(MAKE) update_deps_and_run_tests DIR=service
	$(MAKE) update_deps_and_run_tests DIR=tg-client

update_deps_and_run_tests:
	@echo "Entering directory: $(DIR)"
	cd $(DIR) && \
	go get -u -t ./... && \
	make deps && \
	make ci && \
	cd -

build_and_push_latest:
	$(MAKE) build_and_push DIR=service TAG=$(SERVICE_TAG)
	$(MAKE) build_and_push DIR=tg-client TAG=$(TG_CLIENT_TAG)

build_and_push:
	@echo "Entering directory: $(DIR)"
	cd $(DIR) && \
	make build-linux-amd64 && \
	make build-docker DOCKER_REGISTRY=$(DOCKER_REGISTRY) SERVICE_TAG=$(TAG) && \
	make push-docker DOCKER_REGISTRY=$(DOCKER_REGISTRY) SERVICE_TAG=$(TAG) && \
	cd -

lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	golangci-lint run --allow-parallel-runners -c ../.golangci.yml
