package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"groupie-tracker/internal/groupie"
)

func TestHomeRendersCatalog(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "Groupie Tracker")
	assertBodyContains(t, rec, "Queen")
	assertBodyContains(t, rec, "/artist/1")
}

func TestArtistRendersDetail(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/artist/1", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "Queen")
	assertBodyContains(t, rec, "Freddie Mercury")
	assertBodyContains(t, rec, "london-uk")
	assertBodyContains(t, rec, "14-12-1973")
}

func TestArtistReturns404ForInvalidID(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/artist/not-a-number", newFakeCatalog())

	assertStatus(t, rec, http.StatusNotFound)
	assertBodyContains(t, rec, "Page not found")
}

func TestArtistReturns404ForMissingID(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/artist/999", newFakeCatalog())

	assertStatus(t, rec, http.StatusNotFound)
	assertBodyContains(t, rec, "Page not found")
}

func TestUnknownRouteReturns404(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/missing", newFakeCatalog())

	assertStatus(t, rec, http.StatusNotFound)
	assertBodyContains(t, rec, "Page not found")
}

func TestSearchReturnsJSON(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/api/search?q=queen", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want JSON", contentType)
	}

	var response searchResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if response.Query != "queen" {
		t.Fatalf("query = %q, want queen", response.Query)
	}
	if len(response.Results) != 1 || response.Results[0].Name != "Queen" {
		t.Fatalf("results = %#v, want Queen", response.Results)
	}
}

func TestHealthzReturnsOK(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/healthz", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	if rec.Body.String() != "ok\n" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}

func TestStaticAssetRoute(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/static/site.css", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "body")
}

func TestMethodNotAllowed(t *testing.T) {
	rec := serveTestRequest(http.MethodPost, "/healthz", newFakeCatalog())

	assertStatus(t, rec, http.StatusMethodNotAllowed)
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", allow)
	}
	assertBodyContains(t, rec, "Method not allowed")
}

func TestInternalServerErrorWhenCatalogMissing(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/", nil)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertBodyContains(t, rec, "Internal server error")
}

func TestPanicRecoveryReturns500(t *testing.T) {
	catalog := newFakeCatalog()
	catalog.panicOnArtists = true

	rec := serveTestRequest(http.MethodGet, "/", catalog)

	assertStatus(t, rec, http.StatusInternalServerError)
	assertBodyContains(t, rec, "Internal server error")
}

func serveTestRequest(method string, path string, catalog Catalog) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	New(catalog).Handler().ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, want, rec.Body.String())
	}
}

func assertBodyContains(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	if !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("body missing %q: %s", want, rec.Body.String())
	}
}

type fakeCatalog struct {
	panicOnArtists bool
}

func newFakeCatalog() *fakeCatalog {
	return &fakeCatalog{}
}

func (c *fakeCatalog) Artists() []groupie.ArtistSummary {
	if c.panicOnArtists {
		panic("artist list failed")
	}

	return []groupie.ArtistSummary{
		{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973"},
	}
}

func (c *fakeCatalog) ArtistByID(id int) (groupie.ArtistDetail, bool) {
	if id != 1 {
		return groupie.ArtistDetail{}, false
	}

	return groupie.ArtistDetail{
		ID:           1,
		Name:         "Queen",
		Image:        "queen.jpeg",
		Members:      []string{"Freddie Mercury", "Brian May"},
		CreationDate: 1970,
		FirstAlbum:   "14-12-1973",
		Locations:    []string{"london-uk"},
		Dates:        []string{"14-12-1973"},
		DatesLocations: map[string][]string{
			"london-uk": {"14-12-1973"},
		},
	}, true
}

func (c *fakeCatalog) Search(query string) []groupie.SearchResult {
	if strings.EqualFold(strings.TrimSpace(query), "queen") {
		return []groupie.SearchResult{
			{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973"},
		}
	}
	return nil
}
