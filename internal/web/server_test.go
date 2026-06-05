package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestHomeRendersArtistCards(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "<img src=\"queen.jpeg\"")
	assertBodyContains(t, rec, "Members: Freddie Mercury, Brian May")
	assertBodyContains(t, rec, "Created 1970")
}

func TestHomeHandlesEmptyCatalog(t *testing.T) {
	emptyCatalog := &fakeCatalog{emptyArtists: true}

	rec := serveTestRequest(http.MethodGet, "/", emptyCatalog)

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "No artists available")
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

func TestArtistDetailShowsAuditExamples(t *testing.T) {
	catalog := newAuditCatalog()

	cases := []struct {
		id   int
		want []string
	}{
		{1, []string{"Queen", "Freddie Mercury", "John Daecon", "Roger Meddows-Taylor"}},
		{2, []string{"Gorillaz", "26-03-2001"}},
		{30, []string{"Travis Scott", "santiago-chile", "turku-finland", "05-07-2019"}},
		{4, []string{"Foo Fighters", "Dave Grohl", "Rami Jaffee"}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("id=%d", tc.id), func(t *testing.T) {
			rec := serveTestRequest(http.MethodGet, "/artist/"+strconv.Itoa(tc.id), catalog)
			assertStatus(t, rec, http.StatusOK)
			for _, want := range tc.want {
				assertBodyContains(t, rec, want)
			}
		})
	}
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

func TestStaticCSSIncludesAccessibilityStyles(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/static/site.css", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/css") {
		t.Fatalf("Content-Type = %q, want text/css", contentType)
	}
	assertBodyContains(t, rec, ".catalog-link:focus-visible")
	assertBodyContains(t, rec, "@media (max-width: 560px)")
}

func TestHomeIncludesSearchScript(t *testing.T) {
	rec := serveTestRequest(http.MethodGet, "/", newFakeCatalog())

	assertStatus(t, rec, http.StatusOK)
	assertBodyContains(t, rec, "<script src=\"/static/site.js\"")
}

func TestSearchEmptyAndNoMatch(t *testing.T) {
	// no-match
	rec := serveTestRequest(http.MethodGet, "/api/search?q=__no_such_artist__", newFakeCatalog())
	assertStatus(t, rec, http.StatusOK)
	var resp searchResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected no results, got %#v", resp.Results)
	}

	// empty query returns results
	rec = serveTestRequest(http.MethodGet, "/api/search?q=", newFakeCatalog())
	assertStatus(t, rec, http.StatusOK)
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatalf("expected some results for empty query, got %#v", resp.Results)
	}
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
	emptyArtists   bool
}

func newFakeCatalog() *fakeCatalog {
	return &fakeCatalog{}
}

func (c *fakeCatalog) Artists() []groupie.ArtistSummary {
	if c.panicOnArtists {
		panic("artist list failed")
	}
	if c.emptyArtists {
		return nil
	}

	return []groupie.ArtistSummary{
		{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973", MemberSummary: "Freddie Mercury, Brian May"},
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
	q := strings.TrimSpace(query)
	if q == "" {
		// return all artists as search results
		artists := c.Artists()
		results := make([]groupie.SearchResult, 0, len(artists))
		for _, a := range artists {
			results = append(results, groupie.SearchResult{ID: a.ID, Name: a.Name, Image: a.Image, CreationDate: a.CreationDate, FirstAlbum: a.FirstAlbum})
		}
		return results
	}
	if strings.EqualFold(q, "queen") {
		return []groupie.SearchResult{{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973"}}
	}
	return nil
}

type auditCatalog struct{}

func newAuditCatalog() *auditCatalog {
	return &auditCatalog{}
}

func (c *auditCatalog) Artists() []groupie.ArtistSummary {
	return []groupie.ArtistSummary{
		{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973"},
		{ID: 2, Name: "Gorillaz", Image: "gorillaz.jpeg", CreationDate: 1998, FirstAlbum: "26-03-2001"},
		{ID: 4, Name: "Foo Fighters", Image: "foofighters.jpeg", CreationDate: 1994, FirstAlbum: "04-07-1995"},
		{ID: 30, Name: "Travis Scott", Image: "travis.jpeg", CreationDate: 2008, FirstAlbum: "04-09-2015"},
	}
}

func (c *auditCatalog) ArtistByID(id int) (groupie.ArtistDetail, bool) {
	switch id {
	case 1:
		return groupie.ArtistDetail{
			ID:           1,
			Name:         "Queen",
			Image:        "queen.jpeg",
			Members:      []string{"Freddie Mercury", "Brian May", "John Daecon", "Roger Meddows-Taylor", "Mike Grose", "Barry Mitchell", "Doug Fogie"},
			CreationDate: 1970,
			FirstAlbum:   "14-12-1973",
			Locations:    []string{"london-uk"},
			Dates:        []string{"14-12-1973"},
			DatesLocations: map[string][]string{
				"london-uk": {"14-12-1973"},
			},
		}, true
	case 2:
		return groupie.ArtistDetail{
			ID:           2,
			Name:         "Gorillaz",
			Image:        "gorillaz.jpeg",
			Members:      []string{"Damon Albarn", "Jamie Hewlett"},
			CreationDate: 1998,
			FirstAlbum:   "26-03-2001",
			Locations:    []string{"paris-france"},
			Dates:        []string{"26-03-2001"},
			DatesLocations: map[string][]string{
				"paris-france": {"26-03-2001"},
			},
		}, true
	case 4:
		return groupie.ArtistDetail{
			ID:           4,
			Name:         "Foo Fighters",
			Image:        "foofighters.jpeg",
			Members:      []string{"Dave Grohl", "Nate Mendel", "Taylor Hawkins", "Chris Shiflett", "Pat Smear", "Rami Jaffee"},
			CreationDate: 1994,
			FirstAlbum:   "04-07-1995",
			Locations:    []string{"seattle-usa"},
			Dates:        []string{"04-07-1995"},
			DatesLocations: map[string][]string{
				"seattle-usa": {"04-07-1995"},
			},
		}, true
	case 30:
		return groupie.ArtistDetail{
			ID:           30,
			Name:         "Travis Scott",
			Image:        "travis.jpeg",
			Members:      []string{"Travis Scott"},
			CreationDate: 2008,
			FirstAlbum:   "04-09-2015",
			Locations:    []string{"santiago-chile", "sao_paulo-brazil", "los_angeles-usa", "houston-usa", "atlanta-usa", "new_orleans-usa", "philadelphia-usa", "london-uk", "frauenfeld-switzerland", "turku-finland"},
			Dates:        []string{"05-07-2019"},
			DatesLocations: map[string][]string{
				"turku-finland": {"05-07-2019"},
			},
		}, true
	default:
		return groupie.ArtistDetail{}, false
	}
}

func (c *auditCatalog) Search(query string) []groupie.SearchResult {
	return nil
}
