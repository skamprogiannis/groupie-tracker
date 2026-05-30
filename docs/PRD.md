# Product Requirements Document (PRD) - Groupie Tracker

## 1. Problem Statement

Build a Go website that consumes the Groupie Tracker API and presents artist information clearly for a peer audit. The site must use artist, location, date, and relation data, and it must include a user-triggered client-server event.

## 2. Users

- Zone01 auditors verifying functional requirements.
- Students exploring artists, members, albums, and concert history.
- Team members who need a small, maintainable standard-library Go project.

## 3. Goals

- Display all artists in a user-friendly catalog.
- Show complete artist detail pages with members, first album, creation year, locations, dates, and relation data.
- Provide a client-server search event using the correct HTTP method.
- Keep the server reliable: no crashes, friendly errors, and tested behavior.
- Include bonus async behavior using goroutines and channels.
- Prepare for deployment through `PORT` and `/healthz`.

## 4. Functional Requirements

- Fetch and decode `/artists`, `/locations`, `/dates`, and `/relation`.
- Join the four API resources into detail models.
- Render a home page with artist cards.
- Render `/artist/{id}` pages for individual artists.
- Provide `GET /api/search?q=...` returning JSON results.
- Update the catalog page with search results without a full page reload.
- Serve static CSS and JavaScript through the Go server.
- Return 404 for unknown routes and invalid artist IDs.
- Return 500 for internal failures without crashing.

## 5. Audit Cases

- Queen members must be visible.
- Gorillaz first album must show `26-03-2001`.
- Travis Scott locations must be visible, including `santiago-chile`, `sao_paulo-brasil`, `los_angeles-usa`, `houston-usa`, `atlanta-usa`, `new_orleans-usa`, `philadelphia-usa`, `london-uk`, `frauenfeld-switzerland`, and `turku-finland`.
- Foo Fighters members must be visible.
- A client action must trigger a server request-response flow.
- The server must use the right HTTP method for the action.
- Pages must work without unexpected 404s.
- HTTP 500 must be handled.

## 6. Non-Functional Requirements

- Go standard library only.
- `go test ./...` must pass.
- Code must be formatted with `gofmt`.
- The website must be responsive and readable on mobile and desktop.
- API failures and malformed requests must not crash the process.

## 7. Out Of Scope

- User accounts.
- Persistent database storage.
- Non-standard Go routers or frontend frameworks.
- Complex maps or charting libraries.

## 8. Milestones

1. Planning and skeleton.
2. Data and server foundation.
3. Required website experience.
4. Bonus, QA, deployment, and release.
