#!/bin/sh

make deps

docker compose -f lale.yml build --parallel
docker compose -f lale.yml up --remove-orphans --force-recreate
