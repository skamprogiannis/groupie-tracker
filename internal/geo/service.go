package geo

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// Service resolves location slugs to coordinates. Every slug is geocoded at
// most once: results (and failures) are cached, and live geocoding calls are
// serialized and rate-limited to respect the provider's usage policy.
type Service struct {
	geocoder Geocoder
	minDelay time.Duration

	mu     sync.Mutex
	cache  map[string]Coord
	misses map[string]bool

	liveMu sync.Mutex // serializes live geocoding calls
	last   time.Time
}

// NewService builds a Service around a geocoder.
func NewService(geocoder Geocoder) *Service {
	return &Service{
		geocoder: geocoder,
		minDelay: time.Second, // Nominatim asks for at most 1 request/second
		cache:    make(map[string]Coord),
		misses:   make(map[string]bool),
	}
}

// Seed preloads coordinates, e.g. from a committed cache file, so common
// locations resolve instantly without a network call.
func (s *Service) Seed(coords map[string]Coord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for slug, coord := range coords {
		s.cache[normalizeSlug(slug)] = coord
	}
}

// SeedJSON preloads coordinates from JSON of the form {"slug": {"lat":..,"lon":..}}.
func (s *Service) SeedJSON(data []byte) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	var coords map[string]Coord
	if err := json.Unmarshal(data, &coords); err != nil {
		return err
	}
	s.Seed(coords)
	return nil
}

// Locate returns the coordinate for a "city-country" slug. ok is false when the
// location is not cached and cannot be geocoded (network error, no match, or a
// cancelled context).
func (s *Service) Locate(ctx context.Context, slug string) (Coord, bool) {
	slug = normalizeSlug(slug)
	if coord, ok := s.cached(slug); ok {
		return coord, true
	}
	if s.isMiss(slug) {
		return Coord{}, false
	}

	// Serialize live calls so the rate limit is honored across goroutines.
	s.liveMu.Lock()
	defer s.liveMu.Unlock()

	// Another goroutine may have resolved it while we waited for the lock.
	if coord, ok := s.cached(slug); ok {
		return coord, true
	}
	if s.isMiss(slug) {
		return Coord{}, false
	}

	if wait := s.minDelay - time.Since(s.last); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return Coord{}, false
		}
	}

	coord, err := s.geocoder.Geocode(ctx, SlugToQuery(slug))
	s.last = time.Now()
	if err != nil {
		s.markMiss(slug)
		return Coord{}, false
	}
	s.store(slug, coord)
	return coord, true
}

func (s *Service) cached(slug string) (Coord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	coord, ok := s.cache[slug]
	return coord, ok
}

func (s *Service) isMiss(slug string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.misses[slug]
}

func (s *Service) store(slug string, coord Coord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[slug] = coord
}

func (s *Service) markMiss(slug string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.misses[slug] = true
}

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
