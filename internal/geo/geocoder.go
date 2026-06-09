// Package geo converts concert location slugs ("city-country") into geographic
// coordinates using a geocoding service, and caches the results.
package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Coord is a geographic coordinate.
type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Geocoder converts a free-form place query (e.g. "mainz, germany") into a
// coordinate.
type Geocoder interface {
	Geocode(ctx context.Context, query string) (Coord, error)
}

const defaultNominatimEndpoint = "https://nominatim.openstreetmap.org/search"

// NominatimGeocoder calls the OpenStreetMap Nominatim geocoding service. Only
// the Go standard library is used.
type NominatimGeocoder struct {
	Client    *http.Client
	Endpoint  string
	UserAgent string
}

// NewNominatimGeocoder returns a geocoder with sensible defaults. Nominatim
// requires an identifying User-Agent.
func NewNominatimGeocoder() *NominatimGeocoder {
	return &NominatimGeocoder{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Endpoint:  defaultNominatimEndpoint,
		UserAgent: "groupie-tracker-geolocalization/1.0 (Zone01 student project)",
	}
}

// Geocode returns the best-match coordinate for query.
func (g *NominatimGeocoder) Geocode(ctx context.Context, query string) (Coord, error) {
	endpoint := g.Endpoint
	if endpoint == "" {
		endpoint = defaultNominatimEndpoint
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return Coord{}, err
	}
	req.Header.Set("User-Agent", g.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return Coord{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Coord{}, fmt.Errorf("geocode %q: unexpected status %d", query, resp.StatusCode)
	}

	var matches []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return Coord{}, fmt.Errorf("geocode %q: decode response: %w", query, err)
	}
	if len(matches) == 0 {
		return Coord{}, fmt.Errorf("geocode %q: no match", query)
	}

	lat, errLat := strconv.ParseFloat(matches[0].Lat, 64)
	lon, errLon := strconv.ParseFloat(matches[0].Lon, 64)
	if errLat != nil || errLon != nil {
		return Coord{}, fmt.Errorf("geocode %q: invalid coordinates", query)
	}
	return Coord{Lat: lat, Lon: lon}, nil
}

// SlugToQuery turns a "city-country" slug into a human geocoding query, e.g.
// "north_carolina-usa" -> "north carolina, usa".
func SlugToQuery(slug string) string {
	city, country, found := strings.Cut(strings.TrimSpace(slug), "-")
	city = strings.ReplaceAll(city, "_", " ")
	if !found {
		return city
	}
	country = strings.ReplaceAll(country, "_", " ")
	return city + ", " + country
}
