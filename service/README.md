# lale-service

gRPC backend for the [Lale](../README.md) spaced-repetition flashcard system. It owns the card lifecycle (create, inspect, update, learn, repeat, delete), schedules reviews using a spaced-repetition algorithm, and enriches cards with dictionary entries, AI-generated content, and TTS audio.

## What it does

- **Card CRUD** — `CreateCard`, `UpdateCard`, `DeleteCard`, `GetAllCards`, `InspectCard`
- **Spaced repetition** — `UpdateCardPerformance` advances the schedule; `GetCardsToLearn` / `GetCardsToRepeat` return the due queues; `MarkCardLearnt` retires a card
- **AI helpers** — `PromptCard` (family-word translations), `GetSentences` (example usage), `GenerateStory` (cohesive paragraph from a user's vocabulary)
- **Audio** — words are pronounced in en-GB, en-US, and en-AU via Google Cloud TTS at creation time

The full gRPC contract is in [`api/lale-service.proto`](api/lale-service.proto).

## Architecture

```
cmd/service             — entrypoint, wires dependencies, starts gRPC + infra servers
internal/grpc           — gRPC handlers and request/response transformers
internal/core           — business logic (validation, session, card workflows)
internal/algo           — spaced-repetition scheduling
internal/repo/card      — MongoDB-backed card repository
internal/repo/dictionary — dictionary client (with stub fallback)
internal/repo/chatgpt   — OpenAI/ChatGPT client
internal/repo/session   — in-memory user-session lock
internal/infrastructure — auxiliary HTTP server (Prometheus /metrics + pprof)
internal/observability  — Mongo command-monitor metrics
pkg/                    — reusable building blocks (entity, logger, speech, openai, gracefulmongo, future)
```

## Configuration

All configuration is read from environment variables via `envconfig`.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `APP_GRPC_PORT` | yes | — | gRPC listen port |
| `APP_INFRA_SERVER_PORT` | no | `8888` | HTTP port for `/metrics` and pprof |
| `APP_LOG_LEVEL` | yes | — | logrus level (`debug`, `info`, …) |
| `APP_MONGO_CARD_PROTOCOL` | yes | — | e.g. `mongodb` or `mongodb+srv` |
| `APP_MONGO_CARD_HOST` | yes | — | MongoDB host |
| `APP_MONGO_CARD_PORT` | no | — | MongoDB port |
| `APP_MONGO_CARD_URI_PARAMS` | yes | — | URI params map (e.g. `retryWrites=true,w=majority`) |
| `APP_MONGO_CARD_DATABASE` | yes | — | Database name |
| `APP_MONGO_CARD_COLLECTION` | yes | — | Collection name |
| `APP_MONGO_CARD_MAX_POOL_SIZE` | no | `100` | Connection pool size |
| `APP_MONGO_USER` / `APP_MONGO_PASS` | yes | — | MongoDB credentials |
| `APP_DICTIONARY_HOST` | no | — | Dictionary service host; if empty the stub is used |
| `APP_DICTIONARY_RETRIES` | no | `3` | Dictionary retry count |
| `APP_DICTIONARY_TIMEOUT` | no | `5s` | Dictionary timeout |
| `APP_DICTIONARY_STUB_ENABLED` | no | `false` | Force the dictionary stub |
| `APP_OPENAI_ADDR` | no | OpenAI API URL | OpenAI endpoint |
| `APP_OPENAI_TOKEN` | no | — | OpenAI bearer token |
| `APP_OPENAI_RETRIES` | no | `3` | OpenAI retry count |
| `APP_OPENAI_TIMEOUT` | no | `30s` | OpenAI timeout |
| `APP_OPENAI_STUB_ENABLED` | no | `false` | Use OpenAI stub |
| `APP_GOOGLE_PROJECT_KEY_JSON` | no | — | Google Cloud service-account JSON for TTS |
| `APP_GOOGLE_STUB_ENABLED` | no | `false` | Use TTS stub |

## Build & run

```sh
make deps          # tidy + verify
make ci            # lint + test
make build         # local-host binary at bin/svc
make build-linux-amd64
make build-docker  # builds ghcr.io/genvmoroz/lale-service:development by default
```

### Tests, benchmarks, coverage

```sh
make test
make bench
make test_coverage
```

### Regenerating gRPC stubs

```sh
make gen           # also runs go generate ./...
make protoc-gen    # just the .proto compile step
```

### Vulnerability scan

```sh
make vulnerabilities_lookup
```

## Observability

- `:${APP_INFRA_SERVER_PORT}/metrics` — Prometheus scrape endpoint (includes `mongo_client_*` operation counters and latency histograms)
- `:${APP_INFRA_SERVER_PORT}/debug/pprof/*` — standard pprof endpoints

## Image

`ghcr.io/genvmoroz/lale-service:latest`
