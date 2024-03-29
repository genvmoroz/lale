.PHONY: run-infra
run-infra:
	docker compose -f ./deployments/docker-compose.yml up -d --remove-orphans

.PHONY: stop-infra
stop-infra:
	docker compose -f ./deployments/docker-compose.yml down

.PHONY: run-locally
run-locally:
#  TODO: implement it
