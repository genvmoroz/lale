#!/bin/sh

.PHONY: run_full
run_full:
	make start_containers
	echo "Waiting for the server to start..."
	sleep 5
	make run_loader
	make clean_containers

.PHONY: run_loader
run_loader:
	go run ./loader/main.go \
		--host=localhost \
		--port=8080 \
		--users=100 \
		--cards-per-user=100 \
		--words-per-card=10 \
		--create-card=true \
        --delete-card=true \
        --generate-story=true \
        --get-all-cards=true \
        --get-sentences=true \
        --get-to-learn=true \
        --get-to-repeat=true \
        --inspect-card=true \
        --prompt-card=true \
        --update-card=true \
        --update-performance=true

.PHONY: build
build:
	go build -o ./loader/bin/loader ./loader/.

.PHONY: start_containers
start_containers: clean_containers
	docker compose -f ./docker-compose.yml up --build -d

.PHONY: clean_containers
clean_containers:
	docker compose -f ./docker-compose.yml down --remove-orphans
