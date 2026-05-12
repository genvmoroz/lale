# lale-tg-client

Telegram bot frontend for the [Lale](../README.md) spaced-repetition flashcard system. It speaks the Telegram Bot API on one side and the [`lale-service`](../service) gRPC API on the other, exposing card creation, review, and inspection as a chat-driven state machine.

## Bot states

Each state is a self-contained conversational flow under [`internal/state/`](internal/state):

| State | Purpose |
| ----- | ------- |
| `create`   | Walk the user through adding a new card |
| `inspect`  | Show details for a single card or word |
| `getall`   | List all cards for the user |
| `update`   | Edit an existing card |
| `learn`    | Drill cards that are due for first-time learning |
| `repeat`   | Drill cards that are due for repetition |
| `story`    | Generate an AI story over the user's vocabulary |
| `learnt`   | Mark a card as fully learnt |
| `help`     | Reference of available commands |

States are wired into the bot in [`cmd/service/main.go`](cmd/service/main.go) via the [`bot-engine`](https://github.com/genvmoroz/bot-engine) dispatcher.

## Configuration

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `APP_LOG_LEVEL` | yes | — | logrus level |
| `APP_TELEGRAM_TOKEN` | yes | — | Telegram bot API token |
| `APP_TELEGRAM_UPDATE_TIMEOUT` | no | `60` | Long-polling timeout (seconds) |
| `APP_LALE_SERVICE_HOST` | yes | — | `lale-service` gRPC host |
| `APP_LALE_SERVICE_PORT` | yes | — | `lale-service` gRPC port |
| `APP_LALE_SERVICE_TIMEOUT` | no | `30s` | Per-call timeout |

## Build & run

```sh
make deps          # tidy + verify
make ci            # tests (no lint target in this module)
make build         # local-host binary at bin/svc
make build-linux-amd64
make build-docker  # ghcr.io/genvmoroz/lale-tg-client:development
```

### Local infra

A docker-compose stack lives under [`deployments/`](deployments) for running the bot together with its dependencies:

```sh
make run-infra     # docker compose up
make stop-infra    # docker compose down
make run-locally   # run the already-built binary
```

### Vulnerability scan

```sh
make vulnerabilities_lookup
```

## Image

`ghcr.io/genvmoroz/lale-tg-client:latest`
