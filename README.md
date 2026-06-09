# Groupie Tracker

**Live site: https://groupietracker-wunu.onrender.com/**

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

- `/` - catalog home page with search, filters, and sorting
- `/artist/{id}` - artist detail page
- `/api/search?q=...` - JSON search endpoint (also accepts `sort`, `minYear`, `maxYear`, `minMembers`, `maxMembers`, `country`)
- `/api/geo?id=...` - JSON concert coordinates for the geolocalization map
- `/healthz` - health check
- `/static/*` - embedded static assets

## Client-server event (async search and filters)

The required client-server event is search. Typing in the search box, choosing a
sort order, or adjusting a filter triggers a `GET /api/search` request from the
browser, and the server answers with JSON that the page renders without a full
reload. The same endpoint powers the autocomplete dropdown, the filtered catalog
grid, and the live "showing N of M" count.

Filtering and sorting are applied server-side (in `internal/groupie`), so the
logic is one tested source of truth: free-text match, creation-year range,
member-count range, concert country, and sort by name, newest, oldest, or member
count.

## Geolocalization (map)

Each artist page shows a map of that band's concerts (the Groupie Tracker
Geolocalization project):

- `internal/geo` geocodes each `city-country` location into coordinates using
  the OpenStreetMap **Nominatim** service over `net/http` (standard library
  only). Results are cached in memory so each place is geocoded at most once,
  and a committed seed cache (`internal/geo/seed.json`, embedded) means the
  deployed app resolves the known locations instantly without live calls.
- The browser requests `GET /api/geo?id=<artist>`, which returns the located
  concerts **ordered by concert date**.
- A **Leaflet + OpenStreetMap** map (loaded from a CDN) drops a marker per
  concert, with a popup of the place and dates, and connects them with a dashed
  **date-ordered path** to trace the tour.

As with any map project, the map tiles, the Leaflet library, and the geocoding
service run in the browser / over HTTP; the Go backend itself uses only the
standard library.

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

### Hosting on Render (free)

The repository includes a small multi-stage `Dockerfile`. To deploy:

1. Push the repository to a GitHub repo Render can read.
2. Render dashboard → **New + → Web Service** → connect the repo.
3. Render auto-detects the `Dockerfile`; pick the **Free** instance type and set
   **Health Check Path** to `/healthz`.
4. Render injects `PORT`, which the app already reads, so no extra config is
   needed.

The free tier sleeps after ~15 minutes idle (a ~40s cold start on the next
hit); a scheduled ping to `/healthz` (e.g. cron-job.org every 10 minutes) keeps
it warm.

**Live URL:** https://groupietracker-wunu.onrender.com/ (health check: https://groupietracker-wunu.onrender.com/healthz)

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
