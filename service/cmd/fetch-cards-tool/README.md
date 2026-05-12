# fetch-cards-tool

Interactive CLI that connects to a running [`lale-service`](../../) over gRPC, fetches every card for a user, and reports useful diagnostics.

## What it does

- Prompts for the Lale service host/port and user ID interactively
- Lists every unlearnt card (those with a zero `NextDueDate`) along with their words and IDs
- Flags duplicates — pairs of cards that share at least one word

Useful as a quick read-only inspection tool against a live deployment, e.g. during data cleanup.

## Build & run

```sh
go run .
# or
go build -o bin/fetch-cards-tool . && ./bin/fetch-cards-tool
```

This is its own Go module (`go.mod` in this directory), so it builds independently of the parent `service` module.
