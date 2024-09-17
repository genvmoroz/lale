.PHONY: deps
deps:
	cd service ; go get -u -t ./... ; make deps ; make ci
	cd tg-client ; go get -u -t ./... ; make deps ; make ci
