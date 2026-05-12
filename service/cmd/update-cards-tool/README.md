# update-cards-tool

Interactive CLI that connects to a running [`lale-service`](../../) over gRPC and walks the operator through editing existing cards via the public API.

## What it does

- Prompts for the Lale service host/port and user context
- Fetches the user's cards and lets you pick one to modify
- Submits the change through `UpdateCard` so the service still runs validation, dictionary enrichment, and audio regeneration

Use this when you want a guided, API-level edit rather than touching MongoDB directly. For a low-level bulk script that bypasses the service, see [`update-cards-from-mongo-script`](../update-cards-from-mongo-script).

## Build & run

```sh
go run .
# or
go build -o bin/update-cards-tool . && ./bin/update-cards-tool
```

This is its own Go module (`go.mod` in this directory), so it builds independently of the parent `service` module.
