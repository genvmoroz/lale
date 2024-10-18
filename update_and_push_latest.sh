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

echo "Updating dependencies and running tests"
update_deps_and_run_tests "service"
update_deps_and_run_tests "tg-client"

echo "Building and pushing docker images"
echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"

echo "SERVICE_TAG: $SERVICE_TAG"
docker build -t "$DOCKER_REGISTRY/$SERVICE_TAG" service
docker push "$DOCKER_REGISTRY/$SERVICE_TAG"

echo "TG_CLIENT_TAG: $TG_CLIENT_TAG"
docker build -t "$DOCKER_REGISTRY/$TG_CLIENT_TAG" tg-client
docker push "$DOCKER_REGISTRY/$TG_CLIENT_TAG"
