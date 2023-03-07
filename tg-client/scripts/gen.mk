.PHONY: gen
gen: mockgen

.PHONY: mockgen
mockgen:
	go install github.com/golang/mock/mockgen@latest
	go generate ./...
