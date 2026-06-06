# Groupie Tracker

Groupie Tracker is a Zone01 Go web project that will fetch the public Groupie Tracker API and display artist, member, album, location, date, and relation data in a user-friendly website.

The current implementation includes the data and server foundation: API fetching, normalized in-memory catalog data, basic route rendering, search JSON, static assets, health checks, and shared error handling.

## Run

```bash
go run .
```

The server listens on `http://localhost:8080` by default. Set `PORT` to use another port:

```bash
PORT=3000 go run .
```

## Test

```bash
go test ./...
gofmt -w .
```

The project uses only Go standard library packages.

## Routes

- `/` - catalog home page
- `/artist/{id}` - artist detail page
- `/api/search?q=...` - JSON search endpoint
- `/healthz` - health check
- `/static/*` - embedded static assets

## Client-server event (async search)

The required client-server event is search. Typing in the search box (or
submitting the form) triggers a `GET /api/search?q=...` request from the
browser, and the server answers with JSON that the page renders without a full
reload.

On the server the query is handled asynchronously:

- The HTTP handler wraps each request in a context with a timeout and hands the
  query to a `SearchWorker` (`internal/groupie/worker.go`).
- The worker runs on its own goroutine and communicates only through channels,
  serializing access to the immutable in-memory catalog.
- Each request carries its own context, so a slow or abandoned request is
  cancelled (timeout or client disconnect) instead of blocking.
- On shutdown the server stops accepting requests and the worker goroutine is
  closed cleanly, so there are no leaked goroutines.

## Deployment

The server is deployment-ready:

- It binds to the `PORT` environment variable (defaulting to `8080`), which is
  what most hosts inject.
- `/healthz` returns `200 ok` for health checks and readiness probes.
- All assets (CSS/JS) are embedded in the binary via `embed`, so a single
  compiled binary is fully self-contained.

To run a production-style build locally:

```bash
go build -o groupie-tracker .
PORT=8080 ./groupie-tracker
```

A specific host (e.g. Fly.io, Render, Railway) has not been chosen yet; when
the site is deployed, record the public URL here.

## Audit notes

- Uses artist, location, date, and relation data (joined per artist).
- Audit examples are covered by tests: Queen members, Gorillaz first album
  `26-03-2001`, Travis Scott locations, Foo Fighters members.
- The client-server event uses the correct HTTP method (`GET`); other methods
  return `405` with an `Allow` header.
- Unknown routes and invalid/missing artist IDs return a friendly `404`.
- Upstream API failure, template failure, and malformed input return `500` (or
  an empty result) without crashing; a panic recovery wraps every handler.
- Standard library only — `go.mod` has no third-party dependencies.

Verify locally with:

```bash
go vet ./...
go test -race ./...
gofmt -l .
```

## Documentation

- `docs/PRD.md` - product requirements and audit scope
- `docs/architecture.md` - package, routing, and async design
- `docs/instructions.txt` - original subject instructions
- `docs/audit-questions.txt` - audit checklist
- `docs/llm-log.txt` - LLM usage log
