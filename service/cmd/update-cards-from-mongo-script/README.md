# update-cards-from-mongo-script

One-off migration script that talks directly to MongoDB — **bypassing** [`lale-service`](../../) — to perform bulk transformations on stored cards. Use it for schema migrations or batch fixes that the public gRPC API cannot express.

## What it does

- Connects to the `dictionary` MongoDB database (host/credentials are currently hard-coded in `main.go`; edit before running)
- Loads every card via the local `mongo.Repo`
- Applies an in-script `updateCard` transformation function to each card
- Saves the result back in batches, with a `clinch` task spinner per card

The current `updateCard` is a no-op placeholder; edit `main.go` to define the transformation before each run.

## Build & run

```sh
go run .
```

This is its own Go module (`go.mod` in this directory). Because it manipulates production data directly, treat it as a manual, single-purpose script — review the diff, point it at a backup or staging cluster first, and re-verify before pointing it at the real database.
