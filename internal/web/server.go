package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"groupie-tracker/internal/geo"
	"groupie-tracker/internal/groupie"
)

//go:embed static
var staticFiles embed.FS

// searchTimeout bounds how long a single /api/search request may wait on the
// async search worker before the handler gives up and returns an error.
const searchTimeout = 5 * time.Second

// Catalog supplies the read-only artist data the HTML pages render.
type Catalog interface {
	Artists() []groupie.ArtistSummary
	ArtistByID(id int) (groupie.ArtistDetail, bool)
	Facets() groupie.FilterOptions
}

// Searcher handles the asynchronous client-server search event. The concrete
// implementation (groupie.SearchWorker) runs on its own goroutine and honors
// the request context for timeout and cancellation.
type Searcher interface {
	Search(ctx context.Context, query groupie.SearchQuery) ([]groupie.SearchResult, error)
}

// Locator resolves a "city-country" concert location slug into coordinates for
// the geolocalization map. ok is false when the location cannot be geocoded.
type Locator interface {
	Locate(ctx context.Context, slug string) (geo.Coord, bool)
}

type Server struct {
	catalog       Catalog
	searcher      Searcher
	locator       Locator
	staticHandler http.Handler
}

func New(catalog Catalog, searcher Searcher, locator Locator) *Server {
	return &Server{
		catalog:       catalog,
		searcher:      searcher,
		locator:       locator,
		staticHandler: newStaticHandler(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/artist/", s.artist)
	mux.HandleFunc("/api/search", s.search)
	mux.HandleFunc("/api/geo", s.geo)
	mux.HandleFunc("/healthz", s.healthz)
	mux.Handle("/static/", http.StripPrefix("/static/", s.staticHandler))
	return s.recoverPanic(mux)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.notFound(w, r)
		return
	}
	if !s.allowOnlyGet(w, r) {
		return
	}
	if s.catalog == nil {
		s.internalServerError(w, r)
		return
	}

	artists := s.catalog.Artists()
	facets := s.catalog.Facets()
	// Match the default sort control ("name") for the initial render.
	sort.SliceStable(artists, func(i, j int) bool {
		return strings.ToLower(artists[i].Name) < strings.ToLower(artists[j].Name)
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	writeHead(w, "Groupie Tracker", "")
	fmt.Fprintf(w, `<header class="page-header"><h1>Groupie Tracker</h1><p class="page-subtitle">%s</p></header>`, html.EscapeString(datasetSummary(len(artists), facets)))

	fmt.Fprint(w, `<form id="searchForm" action="/api/search" method="get" role="search"><div id="search-container"><input id="q" name="q" type="search" placeholder="Search artists..." aria-label="Search artists" role="combobox" aria-expanded="false" aria-controls="dropdown" aria-autocomplete="list" autocomplete="off"><div id="dropdown" class="autocomplete-list" role="listbox" aria-label="Artist search results"></div></div><button type="submit">Search</button></form>`)

	writeControls(w, facets)

	fmt.Fprintf(w, `<p class="results-meta" id="result-count" aria-live="polite">Showing %d of %d artists</p>`, len(artists), len(artists))
	fmt.Fprint(w, `<section aria-label="Artists"><ul class="catalog" id="catalog">`)
	for _, artist := range artists {
		writeArtistCard(w, artist.ID, artist.Name, artist.Image, artist.CreationDate, artist.MemberCount, artist.LocationCount)
	}
	fmt.Fprint(w, `</ul><p class="empty-state" id="empty-state" hidden>No artists match these filters.</p></section>`)
	writeFoot(w)
}

func (s *Server) artist(w http.ResponseWriter, r *http.Request) {
	if !s.allowOnlyGet(w, r) {
		return
	}

	id, ok := artistIDFromPath(r.URL.Path)
	if !ok {
		s.notFound(w, r)
		return
	}
	if s.catalog == nil {
		s.internalServerError(w, r)
		return
	}
	artist, found := s.catalog.ArtistByID(id)
	if !found {
		s.notFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	writeHead(w, artist.Name+" - Groupie Tracker", leafletHead)
	fmt.Fprint(w, `<a class="back-button" href="/">&larr; Back to catalog</a>`)

	fmt.Fprintf(w, `<article class="artist"><header class="artist__header"><img class="artist__image" src="%s" alt="%s"><div class="artist__intro"><h1>%s</h1><dl class="facts">`,
		html.EscapeString(artist.Image), html.EscapeString(artist.Name), html.EscapeString(artist.Name))
	writeFact(w, "Formed", strconv.Itoa(artist.CreationDate))
	writeFact(w, "First album", strings.TrimPrefix(artist.FirstAlbum, "*"))
	writeFact(w, "Members", strconv.Itoa(len(artist.Members)))
	writeFact(w, "Cities played", strconv.Itoa(len(artist.Locations)))
	fmt.Fprint(w, `</dl></div></header>`)

	writeMembers(w, artist.Members)
	writeMap(w, artist)
	writeConcerts(w, artist.Locations, artist.DatesLocations)

	fmt.Fprint(w, `</article>`)
	writeFoot(w)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/search" {
		s.notFound(w, r)
		return
	}
	if !s.allowOnlyGet(w, r) {
		return
	}
	if s.searcher == nil {
		s.internalServerError(w, r)
		return
	}

	params := r.URL.Query()
	query := groupie.SearchQuery{
		Text:         params.Get("q"),
		MinYear:      atoiOrZero(params.Get("minYear")),
		MaxYear:      atoiOrZero(params.Get("maxYear")),
		MinAlbumYear: atoiOrZero(params.Get("minAlbumYear")),
		MaxAlbumYear: atoiOrZero(params.Get("maxAlbumYear")),
		MinMembers:   atoiOrZero(params.Get("minMembers")),
		MaxMembers:   atoiOrZero(params.Get("maxMembers")),
		Country:      params.Get("country"),
		Locations:    params["location"],
		Sort:         params.Get("sort"),
	}

	// Hand the query to the async worker with a bounded context. The worker
	// replies on a channel; if it does not answer before the deadline (or the
	// client disconnects), the request is cancelled instead of hanging.
	ctx, cancel := context.WithTimeout(r.Context(), searchTimeout)
	defer cancel()

	results, err := s.searcher.Search(ctx, query)
	if err != nil {
		s.internalServerError(w, r)
		return
	}

	total := 0
	if s.catalog != nil {
		total = len(s.catalog.Artists())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(searchResponse{
		Query:   strings.TrimSpace(query.Text),
		Total:   total,
		Count:   len(results),
		Results: results,
	})
}

// geoTimeout bounds a single /api/geo request. With the seeded cache nearly
// every location resolves instantly; the budget covers the rare live geocode.
const geoTimeout = 15 * time.Second

type geoPoint struct {
	Location string    `json:"location"`
	Slug     string    `json:"slug"`
	Lat      float64   `json:"lat"`
	Lon      float64   `json:"lon"`
	Dates    []string  `json:"dates"`
	earliest time.Time // used only for date ordering; not serialized
}

type geoResponse struct {
	Artist string     `json:"artist"`
	Points []geoPoint `json:"points"`
}

// geo returns an artist's concert locations as map points, ordered by concert
// date so the client can draw the tour path. This is the geolocalization
// client-server event.
func (s *Server) geo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/geo" {
		s.notFound(w, r)
		return
	}
	if !s.allowOnlyGet(w, r) {
		return
	}
	if s.catalog == nil || s.locator == nil {
		s.internalServerError(w, r)
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil || id <= 0 {
		s.notFound(w, r)
		return
	}
	artist, found := s.catalog.ArtistByID(id)
	if !found {
		s.notFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), geoTimeout)
	defer cancel()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(geoResponse{
		Artist: artist.Name,
		Points: s.locateConcerts(ctx, artist),
	})
}

func (s *Server) locateConcerts(ctx context.Context, artist groupie.ArtistDetail) []geoPoint {
	points := make([]geoPoint, 0, len(artist.Locations))
	for _, slug := range artist.Locations {
		coord, ok := s.locator.Locate(ctx, slug)
		if !ok {
			continue
		}
		dates := cleanDates(artist.DatesLocations[slug])
		points = append(points, geoPoint{
			Location: groupie.FormatLocation(slug),
			Slug:     slug,
			Lat:      coord.Lat,
			Lon:      coord.Lon,
			Dates:    dates,
			earliest: earliestDate(dates),
		})
	}
	sort.SliceStable(points, func(i, j int) bool {
		return points[i].earliest.Before(points[j].earliest)
	})
	return points
}

// earliestDate returns the earliest parseable "DD-MM-YYYY" date, or a far-future
// time so undated locations sort last.
func earliestDate(dates []string) time.Time {
	earliest := time.Time{}
	for _, d := range dates {
		t, err := time.Parse("02-01-2006", strings.TrimSpace(d))
		if err != nil {
			continue
		}
		if earliest.IsZero() || t.Before(earliest) {
			earliest = t
		}
	}
	if earliest.IsZero() {
		return time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return earliest
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if !s.allowOnlyGet(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "ok")
}

func newStaticHandler() http.Handler {
	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(staticRoot))
}

func artistIDFromPath(path string) (int, bool) {
	idPart := strings.TrimPrefix(path, "/artist/")
	if idPart == "" || strings.Contains(idPart, "/") {
		return 0, false
	}

	id, err := strconv.Atoi(idPart)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				s.internalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) allowOnlyGet(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}

	w.Header().Set("Allow", http.MethodGet)
	s.errorPage(w, http.StatusMethodNotAllowed, "Method not allowed", "Use GET for this route.")
	return false
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.errorPage(w, http.StatusNotFound, "Page not found", "The requested Groupie Tracker page does not exist.")
}

func (s *Server) internalServerError(w http.ResponseWriter, r *http.Request) {
	s.errorPage(w, http.StatusInternalServerError, "Internal server error", "The server could not complete this request.")
}

func (s *Server) errorPage(w http.ResponseWriter, status int, title string, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title><link rel="stylesheet" href="/static/site.css"></head><body><main><h1>%s</h1><p>%s</p><p><a href="/">Back to catalog</a></p></main></body></html>`, html.EscapeString(title), html.EscapeString(title), html.EscapeString(message))
}

// leafletHead loads the Leaflet map library and styles. It is injected only on
// pages that render a map. Leaflet is deferred so it executes before site.js.
const leafletHead = `<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"><script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" defer></script>`

// writeHead writes the shared document head, decorative background, and opening
// <main> tag. extraHead injects page-specific <head> markup (e.g. the map
// library) and is trusted server-controlled content.
func writeHead(w http.ResponseWriter, title, extraHead string) {
	_, _ = fmt.Fprintf(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title><link rel="stylesheet" href="/static/site.css">%s<script src="/static/site.js" defer></script><script src="https://unpkg.com/@dotlottie/player-component@2.7.1/dist/dotlottie-player.mjs" type="module"></script></head><body><dotlottie-player src="/static/background-globe-rotating.lottie" background="transparent" speed="1" loop autoplay class="animated-bg" aria-hidden="true"></dotlottie-player><div class="bg-veil" aria-hidden="true"></div><main>`, html.EscapeString(title), extraHead)
}

func writeFoot(w http.ResponseWriter) {
	_, _ = fmt.Fprint(w, `</main></body></html>`)
}

// writeControls renders the filter and sort toolbar, populated from the dataset
// facets so the year and country options always reflect the real data.
func writeControls(w http.ResponseWriter, facets groupie.FilterOptions) {
	// The toggle is only shown on small screens (CSS); it collapses the filter
	// bar so the catalog is visible without scrolling past every control.
	_, _ = fmt.Fprint(w, `<button type="button" id="filterToggle" class="filter-toggle" aria-expanded="false" aria-controls="filterBar">Filters &amp; sort</button>`)
	_, _ = fmt.Fprint(w, `<form id="filterBar" class="filters" aria-label="Filter and sort artists"><div class="filter"><label for="f-sort">Sort</label><select id="f-sort" name="sort"><option value="name">Name A–Z</option><option value="newest">Newest first</option><option value="oldest">Oldest first</option><option value="members">Most members</option></select></div>`)

	// Creation-date range filter.
	_, _ = fmt.Fprintf(w, `<div class="filter"><label for="f-min-year">Created from</label><input id="f-min-year" name="minYear" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div><div class="filter"><label for="f-max-year">Created to</label><input id="f-max-year" name="maxYear" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div>`,
		facets.MinYear, facets.MaxYear, facets.MinYear, facets.MinYear, facets.MaxYear, facets.MaxYear)

	// First-album-date range filter.
	_, _ = fmt.Fprintf(w, `<div class="filter"><label for="f-min-album">Album from</label><input id="f-min-album" name="minAlbumYear" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div><div class="filter"><label for="f-max-album">Album to</label><input id="f-max-album" name="maxAlbumYear" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div>`,
		facets.MinAlbumYear, facets.MaxAlbumYear, facets.MinAlbumYear, facets.MinAlbumYear, facets.MaxAlbumYear, facets.MaxAlbumYear)

	// Member-count range filter.
	_, _ = fmt.Fprintf(w, `<div class="filter"><label for="f-min-members">Min members</label><input id="f-min-members" name="minMembers" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div><div class="filter"><label for="f-max-members">Max members</label><input id="f-max-members" name="maxMembers" type="number" inputmode="numeric" min="%d" max="%d" placeholder="%d"></div>`,
		facets.MinMembers, facets.MaxMembers, facets.MinMembers, facets.MinMembers, facets.MaxMembers, facets.MaxMembers)

	// Country dropdown (a quick coarse location filter).
	_, _ = fmt.Fprint(w, `<div class="filter"><label for="f-country">Country</label><select id="f-country" name="country"><option value="">All countries</option>`)
	for _, country := range facets.Countries {
		_, _ = fmt.Fprintf(w, `<option value="%s">%s</option>`, html.EscapeString(country), html.EscapeString(groupie.FormatLocation(country)))
	}
	_, _ = fmt.Fprint(w, `</select></div>`)

	// Concert-location checkbox filter (multiple selection).
	writeLocationFilter(w, facets.Locations)

	_, _ = fmt.Fprint(w, `<button type="reset" id="f-clear" class="filter-clear">Clear</button></form>`)
}

// writeLocationFilter renders the concert-location checkbox filter: a scrollable
// list of every concert location. Checking one or more narrows the catalog to
// artists who played at any of them.
func writeLocationFilter(w http.ResponseWriter, locations []string) {
	if len(locations) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, `<fieldset class="filter filter--locations"><legend>Concert locations (%d)</legend><input type="search" id="loc-search" class="location-search" placeholder="Type to find a location, e.g. japan" aria-label="Filter the location list"><div class="location-list">`, len(locations))
	for _, slug := range locations {
		_, _ = fmt.Fprintf(w, `<label class="loc-option"><input type="checkbox" name="location" value="%s"><span>%s</span></label>`,
			html.EscapeString(slug), html.EscapeString(groupie.FormatLocation(slug)))
	}
	_, _ = fmt.Fprint(w, `</div></fieldset>`)
}

// writeArtistCard renders one catalog card. The browser mirrors this exact
// markup in site.js so filtered results look identical to the initial page.
func writeArtistCard(w http.ResponseWriter, id int, name, image string, creation, members, locations int) {
	_, _ = fmt.Fprintf(w, `<li class="catalog-item"><a class="catalog-link" href="/artist/%d"><img src="%s" alt="%s" loading="lazy" decoding="async"><div class="catalog-card"><h3>%s</h3><p class="catalog-card__meta">Since %d</p><ul class="card-stats"><li>%s</li><li>%s</li></ul></div></a></li>`,
		id, html.EscapeString(image), html.EscapeString(name), html.EscapeString(name), creation, html.EscapeString(countLabel(members, "member")), html.EscapeString(countLabel(locations, "location")))
}

func writeFact(w http.ResponseWriter, label, value string) {
	_, _ = fmt.Fprintf(w, `<div class="fact"><dt>%s</dt><dd>%s</dd></div>`, html.EscapeString(label), html.EscapeString(value))
}

func writeMembers(w http.ResponseWriter, members []string) {
	if len(members) == 0 {
		return
	}
	_, _ = fmt.Fprint(w, `<section class="panel"><h2>Members</h2><ul class="chips">`)
	for _, member := range members {
		_, _ = fmt.Fprintf(w, `<li class="chip">%s</li>`, html.EscapeString(member))
	}
	_, _ = fmt.Fprint(w, `</ul></section>`)
}

// writeMap renders the geolocalization map container. site.js fills it with
// Leaflet markers fetched from /api/geo?id=<id>, drawing the tour path in date
// order. Empty locations are skipped so the map only shows when there is data.
func writeMap(w http.ResponseWriter, artist groupie.ArtistDetail) {
	if len(artist.Locations) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, `<section class="panel"><h2>Concert map</h2><div id="concert-map" class="concert-map" data-artist-id="%d" aria-label="Map of %s concert locations"></div><p class="map-note muted">Each marker is a geocoded concert location; the dashed line follows the tour in date order.</p></section>`,
		artist.ID, html.EscapeString(artist.Name))
}

// writeConcerts renders one entry per location (so every location stays
// visible), pairing it with the dates from the relation data when present. The
// raw slug is kept in a title attribute for accessibility and exact lookups.
func writeConcerts(w http.ResponseWriter, locations []string, datesLocations map[string][]string) {
	_, _ = fmt.Fprint(w, `<section class="panel"><h2>Concerts</h2>`)
	if len(locations) == 0 {
		_, _ = fmt.Fprint(w, `<p class="muted">No concerts listed.</p></section>`)
		return
	}

	ordered := make([]string, len(locations))
	copy(ordered, locations)
	sort.Strings(ordered)

	_, _ = fmt.Fprint(w, `<ul class="concert-list">`)
	for _, location := range ordered {
		dates := cleanDates(datesLocations[location])
		datesHTML := `<span class="concert__dates muted">No dates announced</span>`
		if len(dates) > 0 {
			datesHTML = `<span class="concert__dates">` + html.EscapeString(strings.Join(dates, " · ")) + `</span>`
		}
		_, _ = fmt.Fprintf(w, `<li class="concert"><span class="concert__place" title="%s">%s</span>%s</li>`,
			html.EscapeString(location), html.EscapeString(groupie.FormatLocation(location)), datesHTML)
	}
	_, _ = fmt.Fprint(w, `</ul></section>`)
}

func cleanDates(dates []string) []string {
	cleaned := make([]string, 0, len(dates))
	for _, date := range dates {
		cleaned = append(cleaned, strings.TrimPrefix(date, "*"))
	}
	return cleaned
}

func datasetSummary(count int, facets groupie.FilterOptions) string {
	if count == 0 {
		return "No artists available"
	}
	if facets.MinYear == 0 || facets.MinYear == facets.MaxYear {
		return fmt.Sprintf("%s in the catalog", countLabel(count, "artist"))
	}
	return fmt.Sprintf("%s · formed %d–%d", countLabel(count, "artist"), facets.MinYear, facets.MaxYear)
}

func countLabel(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func atoiOrZero(value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

type searchResponse struct {
	Query   string                 `json:"query"`
	Total   int                    `json:"total"`
	Count   int                    `json:"count"`
	Results []groupie.SearchResult `json:"results"`
}
