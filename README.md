# Lale

A spaced-repetition flashcard system for language learning. The backend (`service`) stores cards in MongoDB and exposes a gRPC API; the Telegram bot frontend (`tg-client`) is the user-facing channel. Cards are enriched with dictionary definitions, AI-generated example sentences and stories, and TTS audio so a single saved word becomes a rich learning unit.

## Repository layout

| Path | Description |
| ---- | ----------- |
| [`service/`](service) | gRPC backend, spaced-repetition core, MongoDB persistence, AI/TTS/dictionary integrations |
| [`tg-client/`](tg-client) | Telegram bot that drives the backend over gRPC |
| [`service/cmd/service/`](service/cmd/service) | Backend service entrypoint |
| [`service/cmd/fetch-cards-tool/`](service/cmd/fetch-cards-tool) | Interactive CLI to list cards and find duplicates |
| [`service/cmd/update-cards-tool/`](service/cmd/update-cards-tool) | Interactive CLI to edit cards via the gRPC API |
| [`service/cmd/update-cards-from-mongo-script/`](service/cmd/update-cards-from-mongo-script) | One-off MongoDB-direct migration script |
| [`service/test/stress/`](service/test/stress) | Stress-test loader for the backend |

The repository is a Go multi-module workspace; each subdirectory under `service/cmd/` with its own `go.mod` builds independently.

## Prerequisites

- Go 1.26+
- Docker (for `make all`)
- `protoc` (only when regenerating gRPC stubs)
- A running MongoDB instance for the backend

## `make all`

The root `Makefile` orchestrates both modules end-to-end:

```sh
make all
```

This runs, for each of `service` and `tg-client`:

1. `go get -u -t ./...` — refresh dependencies
2. `make deps` — `go mod tidy` and `go mod verify`
3. `make ci` — lint (service only) and tests
4. `make build-linux-amd64` — produce a Linux/amd64 binary
5. `make build-docker` / `make push-docker` — build and push `ghcr.io/genvmoroz/lale-{service,tg-client}:latest`

Run a single module's pipeline with `make -C service ci` or `make -C tg-client ci`.

## Lint

The root `.golangci.yml` is the source of truth for both modules. Run from a module directory:

```sh
make -C service lint
```

## Container images

Published to GHCR:

- `ghcr.io/genvmoroz/lale-service:latest`
- `ghcr.io/genvmoroz/lale-tg-client:latest`
