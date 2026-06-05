package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"groupie-tracker/internal/groupie"
)

//go:embed static
var staticFiles embed.FS

type Catalog interface {
	Artists() []groupie.ArtistSummary
	ArtistByID(id int) (groupie.ArtistDetail, bool)
	Search(query string) []groupie.SearchResult
}

type Server struct {
	catalog       Catalog
	staticHandler http.Handler
}

func New(catalog Catalog) *Server {
	return &Server{
		catalog:       catalog,
		staticHandler: newStaticHandler(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/artist/", s.artist)
	mux.HandleFunc("/api/search", s.search)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Groupie Tracker</title><link rel="stylesheet" href="/static/site.css"></head><body><main><h1>Groupie Tracker</h1><form action="/api/search" method="get"><label for="q">Search artists</label><input id="q" name="q" type="search"><button type="submit">Search</button></form><section><h2>Artists</h2><ul class="catalog">`)
	for _, artist := range artists {
		_, _ = fmt.Fprintf(w, `<li><a href="/artist/%d">%s</a><span>%d</span><span>%s</span></li>`, artist.ID, html.EscapeString(artist.Name), artist.CreationDate, html.EscapeString(artist.FirstAlbum))
	}
	_, _ = fmt.Fprint(w, `</ul></section></main></body></html>`)
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
	_, _ = fmt.Fprintf(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s - Groupie Tracker</title><link rel="stylesheet" href="/static/site.css"></head><body><main><a href="/">Back to catalog</a><article><h1>%s</h1><img src="%s" alt="%s"><dl><dt>Creation date</dt><dd>%d</dd><dt>First album</dt><dd>%s</dd></dl>`, html.EscapeString(artist.Name), html.EscapeString(artist.Name), html.EscapeString(artist.Image), html.EscapeString(artist.Name), artist.CreationDate, html.EscapeString(artist.FirstAlbum))
	writeStringList(w, "Members", artist.Members)
	writeStringList(w, "Locations", artist.Locations)
	writeStringList(w, "Dates", artist.Dates)
	writeRelations(w, artist.DatesLocations)
	_, _ = fmt.Fprint(w, `</article></main></body></html>`)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/search" {
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := searchResponse{
		Query:   strings.TrimSpace(r.URL.Query().Get("q")),
		Results: s.catalog.Search(r.URL.Query().Get("q")),
	}
	_ = json.NewEncoder(w).Encode(response)
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

func writeStringList(w http.ResponseWriter, title string, values []string) {
	_, _ = fmt.Fprintf(w, `<section><h2>%s</h2><ul>`, html.EscapeString(title))
	for _, value := range values {
		_, _ = fmt.Fprintf(w, `<li>%s</li>`, html.EscapeString(value))
	}
	_, _ = fmt.Fprint(w, `</ul></section>`)
}

func writeRelations(w http.ResponseWriter, datesLocations map[string][]string) {
	locations := sortedKeys(datesLocations)
	_, _ = fmt.Fprint(w, `<section><h2>Relations</h2><dl>`)
	for _, location := range locations {
		_, _ = fmt.Fprintf(w, `<dt>%s</dt><dd>%s</dd>`, html.EscapeString(location), html.EscapeString(strings.Join(datesLocations[location], ", ")))
	}
	_, _ = fmt.Fprint(w, `</dl></section>`)
}

func sortedKeys(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type searchResponse struct {
	Query   string                 `json:"query"`
	Results []groupie.SearchResult `json:"results"`
}
