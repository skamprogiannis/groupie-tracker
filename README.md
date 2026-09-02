# Groupie Tracker

**[Live demo](https://groupietracker-wunu.onrender.com/)** ·
**[Health check](https://groupietracker-wunu.onrender.com/healthz)**

Groupie Tracker is a Go web application for exploring artists and their concert
history. It joins the public Groupie Tracker API's artist, location, date, and
relation resources into a searchable catalog with server-side filters,
interactive artist pages, and date-ordered concert maps.

The backend uses only the Go standard library. The browser layer is plain HTML,
CSS, and JavaScript, with Leaflet and OpenStreetMap tiles used for the map.

## Highlights

- Search and autocomplete across artists, members, dates, and locations.
- Filter by formation year, first-album year, member count, country, or one or
  more concert locations; sort without a page reload.
- Explore responsive artist profiles with members, release facts, concert
  dates, and a geolocated tour path.
- Process searches through a goroutine-and-channel worker with request
  cancellation, timeouts, and clean shutdown.
- Recover from bad routes, invalid IDs, upstream failures, and unexpected
  handler panics without taking down the server.
- Ship as a self-contained binary or a small, non-root Docker image, with a
  `/healthz` endpoint for deployment checks.

## Architecture

```text
Groupie Tracker API
        │
        ▼
standard-library HTTP client ──► normalized in-memory catalog
                                      │              │
                                      ▼              ▼
                              HTML page handlers   search worker
                                      │              │
                                      └──────┬───────┘
                                             ▼
                                    browser UI and map
```

- `internal/groupie` owns API decoding, catalog normalization, search,
  filtering, sorting, and the asynchronous search worker.
- `internal/geo` resolves concert-location slugs through a seeded,
  rate-limited Nominatim client and caches results in memory.
- `internal/web` owns routes, embedded templates and assets, JSON endpoints,
  and shared error handling.
- `main` wires dependencies, loads the catalog, starts the HTTP server, and
  coordinates graceful shutdown.

See [`docs/architecture.md`](docs/architecture.md) for the detailed data flow
and design decisions.

## Run locally

Requirements: Go 1.22 or newer and network access to load the public artist API.

```bash
go run .
```

Open `http://localhost:8080`. To choose another port:

```bash
PORT=3000 go run .
```

For a production-style local build:

```bash
go build -o groupie-tracker .
PORT=8080 ./groupie-tracker
```

The included multi-stage `Dockerfile` produces a minimal image that runs as a
non-root user.

## Test

```bash
gofmt -l .
go vet ./...
go test -race ./...
go build ./...
```

The test suite covers API failures, catalog joins, audit examples, filters,
sorting, worker cancellation and concurrency, web routes, JSON responses,
geocoding, error pages, and health checks. GitHub Actions runs the same quality
checks for every push and pull request.

## Routes

| Route | Purpose |
| --- | --- |
| `/` | Searchable and filterable artist catalog |
| `/artist/{id}` | Artist details, concerts, and map |
| `/api/search` | JSON search, filter, and sort results |
| `/api/geo?id={id}` | Date-ordered concert coordinates |
| `/healthz` | Readiness response |
| `/static/*` | Embedded CSS, JavaScript, and visual assets |

## Team and contributions

This was a four-person Zone01 Athens project. Roles below are summarized from
the repository history and the project work log:

- **Stefanos Kamprogiannis (`skamprogiannis`)** — project planning, initial
  scaffold, documentation, audit preparation, and portfolio finalization.
- **`nkountou`** — API models and client, catalog normalization and search,
  server foundation, error handling, and foundational tests.
- **Eleftheria Manola (`emanola`)** — the initial catalog interface,
  responsive and accessible styling, client-side search, and autocomplete.
- **Dimitris Rigas (`DimitrisRgs`)** — asynchronous search and shutdown,
  advanced filters and sorting, Docker deployment, geolocalization, map
  visualization, and integration polish.

The Git history is retained so individual contributions remain attributable.

## Status and limitations

This repository represents the completed Zone01 implementation, including the
Search Bar, Filters, Geolocalization, and Visualizations extensions.

- Artist data is loaded from the upstream API at startup and held in memory;
  there is no database or user account system.
- Known coordinates are embedded, but Leaflet, map tiles, and fallback
  geocoding still depend on external services.
- A sleeping free-tier deployment may need a short cold start before the live
  demo responds.

## Project references

- [`docs/PRD.md`](docs/PRD.md) — requirements and audit scope
- [`docs/architecture.md`](docs/architecture.md) — package and request flow
- [`docs/instructions.txt`](docs/instructions.txt) — original subject
- [`docs/audit-questions.txt`](docs/audit-questions.txt) — audit checklist
- [`docs/llm-log.txt`](docs/llm-log.txt) — required tool-assistance log
