#!/bin/sh

set -e

execute_commands() {
  echo "Entering directory: $1"
  cd "$1" || exit
  echo "Current directory: $(pwd)"
  go get -u -t ./...
  make deps
  make ci
  cd - || exit
}

execute_commands "service"
execute_commands "tg-client"
