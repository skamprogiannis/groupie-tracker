package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/internal/groupie"
)

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
		staticHandler: http.NotFoundHandler(),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/artist/", s.artist)
	mux.HandleFunc("/api/search", s.search)
	mux.HandleFunc("/healthz", s.healthz)
	mux.Handle("/static/", http.StripPrefix("/static/", s.staticHandler))
	return mux
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !allowOnlyGet(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "Groupie Tracker catalog")
}

func (s *Server) artist(w http.ResponseWriter, r *http.Request) {
	if !allowOnlyGet(w, r) {
		return
	}

	id, ok := artistIDFromPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "Artist %d\n", id)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/search" {
		http.NotFound(w, r)
		return
	}
	if !allowOnlyGet(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = fmt.Fprintln(w, `{"results":[]}`)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if !allowOnlyGet(w, r) {
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "ok")
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

func allowOnlyGet(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}

	w.Header().Set("Allow", http.MethodGet)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}
