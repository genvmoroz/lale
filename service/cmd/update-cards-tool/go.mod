module update-cards-tool

go 1.24.0

require (
	github.com/genvmoroz/lale/service v1.0.0
	github.com/liamg/clinch v1.6.6
	github.com/liamg/tml v0.7.0
	github.com/samber/lo v1.51.0
	google.golang.org/grpc v1.75.0
)

require (
	github.com/pkg/term v1.1.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace github.com/genvmoroz/lale/service => ../../
