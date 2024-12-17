#!/bin/sh

DOCKER_REGISTRY="ghcr.io/genvmoroz"
SERVICE_TAG="lale-service:latest"
TG_CLIENT_TAG="lale-tg-client:latest"

set -e

update_deps_and_run_tests() {
  echo "Entering directory: $1"
  cd "$1" || exit
  echo "Current directory: $(pwd)"
  go get -u -t ./...
  make deps
  make ci
  cd - || exit
}

build_and_push_latest() {
  echo "Entering directory: $1"
  cd "$1" || exit
  echo "Current directory: $(pwd)"
  make build-linux-amd64
  make build-docker DOCKER_REGISTRY="$DOCKER_REGISTRY" SERVICE_TAG="$2"
  make push-docker DOCKER_REGISTRY="$DOCKER_REGISTRY" SERVICE_TAG="$2"
  cd - || exit
}

echo "Updating dependencies and running tests"
update_deps_and_run_tests "service"
update_deps_and_run_tests "tg-client"

echo "Building and pushing docker images"
build_and_push_latest "service" $SERVICE_TAG
build_and_push_latest "tg-client" $TG_CLIENT_TAG
