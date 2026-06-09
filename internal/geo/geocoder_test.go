package geo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNominatimGeocoderParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "mainz, germany" {
			t.Errorf("query q = %q, want 'mainz, germany'", got)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"lat":"49.9928617","lon":"8.2472526"}]`))
	}))
	defer server.Close()

	g := &NominatimGeocoder{Client: server.Client(), Endpoint: server.URL, UserAgent: "test/1.0"}
	coord, err := g.Geocode(context.Background(), "mainz, germany")
	if err != nil {
		t.Fatalf("Geocode error: %v", err)
	}
	if coord.Lat != 49.9928617 || coord.Lon != 8.2472526 {
		t.Fatalf("coord = %+v, want lat 49.99.., lon 8.24..", coord)
	}
}

func TestNominatimGeocoderNoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	g := &NominatimGeocoder{Client: server.Client(), Endpoint: server.URL}
	if _, err := g.Geocode(context.Background(), "nowhere-land"); err == nil {
		t.Fatal("expected an error for an empty result, got nil")
	}
}

func TestSlugToQuery(t *testing.T) {
	cases := map[string]string{
		"north_carolina-usa":             "north carolina, usa",
		"abu_dhabi-united_arab_emirates": "abu dhabi, united arab emirates",
		"london-uk":                      "london, uk",
		"singlecity":                     "singlecity",
	}
	for slug, want := range cases {
		if got := SlugToQuery(slug); got != want {
			t.Errorf("SlugToQuery(%q) = %q, want %q", slug, got, want)
		}
	}
}
