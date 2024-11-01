#!/bin/sh

.PHONY: run_stress_test
run_stress_test:
	make run_containers
	#todo: add health check, implement a health check endpoint in the server
	echo "Waiting for the server to start..."
	sleep 5
	make run_loader
	make stop_containers

.PHONY: run_loader
run_loader:
	go run ./loader/main.go \
		--host=localhost \
		--port=8080 \
		--users=10000 \
		--cards-per-user=2 \
		--words-per-card=1 \
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

.PHONY: run_containers
run_containers: stop_containers
	docker compose -f ./docker-compose.yml up -d

.PHONY: stop_containers
stop_containers:
	docker compose -f ./docker-compose.yml down --remove-orphans
