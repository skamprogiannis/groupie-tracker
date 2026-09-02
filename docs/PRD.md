# Product Requirements Document (PRD) - Groupie Tracker

## 1. Problem Statement

Build a Go website that consumes the Groupie Tracker API, manipulates the data,
and displays it in a presentable, responsive visualization-focused interface for
a peer audit. The site must use artist, location, date, and relation data, must
contain CSS, and must keep the base subject's reliability principles.

## 2. Users

- Zone01 auditors verifying functional requirements.
- Students exploring artists, members, albums, and concert history.
- Team members who need a small, maintainable standard-library Go project.

## 3. Goals

- Display all artists in a user-friendly catalog.
- Show complete artist detail pages with members, first album, creation year, locations, dates, and relation data.
- Provide a client-server search event using the correct HTTP method.
- Present API data through clear visualizations: artist cards, fact panels,
  member chips, filter controls, concert lists, and a concert map.
- Follow Schneiderman's 8 Golden Rules of Interface Design: consistency,
  shortcuts, informative feedback, closure, simple error handling, reversal,
  internal locus of control, and reduced short-term memory load.
- Keep the server reliable: no crashes, friendly errors, and tested behavior.
- Include bonus async behavior using goroutines and channels.
- Prepare for deployment through `PORT` and `/healthz`.
- Map each artist's concert locations with geocoded coordinates (Geolocalization extension).

## 4. Functional Requirements

- Fetch and decode `/artists`, `/locations`, `/dates`, and `/relation`.
- Join the four API resources into detail models.
- Render a home page with artist cards.
- Render `/artist/{id}` pages for individual artists.
- Provide `GET /api/search?q=...` returning JSON results, with filter and sort
  query parameters.
- Update the catalog page with search, filter, and sort results without a full
  page reload.
- Geocode concert locations and expose `GET /api/geo?id=...` for the map.
- Serve static CSS and JavaScript through the Go server.
- Provide a consistent, responsive, contrast-conscious CSS theme with visible
  focus states and a background visual treatment.
- Return 404 for unknown routes and invalid artist IDs.
- Return 500 for internal failures without crashing.

## 5. Audit Cases

- Queen members must be visible.
- Gorillaz first album must show `26-03-2001`.
- Travis Scott locations must be visible, including `santiago-chile`,
  `sao_paulo-brazil` (the live API value; some audit sheets spell it
  `sao_paulo-brasil`), `los_angeles-usa`, `houston-usa`, `atlanta-usa`,
  `new_orleans-usa`, `philadelphia-usa`, `london-uk`,
  `frauenfeld-switzerland`, and `turku-finland`.
- Foo Fighters members must be visible.
- A client action must trigger a server request-response flow.
- The server must use the right HTTP method for the action.
- Pages must work without unexpected 404s.
- HTTP 500 must be handled.
- Unknown pages must return a styled 404 page.
- The visual design must be usable, consistent, responsive, and readable.

## 6. Non-Functional Requirements

- Go standard library only.
- `go test ./...` must pass.
- Code must be formatted with `gofmt`.
- The website must be responsive and readable on mobile and desktop.
- API failures and malformed requests must not crash the process.
- The app must run quickly, avoid unnecessary repeated API requests, and keep
  data in an in-memory catalog after startup.

## 7. Out Of Scope

- User accounts.
- Persistent database storage.
- Non-standard Go routers or frontend frameworks.
- Analytics dashboards. The Geolocalization extension adds a concert map via a
  standard browser map library, which is in scope for visualization.

## 8. Milestones

1. Planning and skeleton.
2. Data and server foundation.
3. Required website experience.
4. Bonus, QA, deployment, and release.
