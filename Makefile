.PHONY: deps
deps:
	cd service ; tolatest go.mod ; make deps ; make ci
	cd tg-client ; tolatest go.mod ; make deps ; make ci

