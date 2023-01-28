.PHONY: gen
gen: protoc-gen mockgen

.PHONY: protoc-gen
protoc-gen:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/**.proto

.PHONY: mockgen
mockgen:
	go install github.com/golang/mock/mockgen@latest
	go generate ./...
