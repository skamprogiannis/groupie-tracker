package geo

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

type stubGeocoder struct {
	calls int32
	coord Coord
	err   error
}

func (s *stubGeocoder) Geocode(context.Context, string) (Coord, error) {
	atomic.AddInt32(&s.calls, 1)
	return s.coord, s.err
}

func TestServiceGeocodesAndCaches(t *testing.T) {
	stub := &stubGeocoder{coord: Coord{Lat: 1, Lon: 2}}
	svc := NewService(stub)
	svc.minDelay = 0

	for i := 0; i < 3; i++ {
		coord, ok := svc.Locate(context.Background(), "London-UK")
		if !ok || coord != (Coord{Lat: 1, Lon: 2}) {
			t.Fatalf("Locate = %+v ok=%v, want {1 2} true", coord, ok)
		}
	}
	if got := atomic.LoadInt32(&stub.calls); got != 1 {
		t.Fatalf("geocoder called %d times, want 1 (result cached)", got)
	}
}

func TestServiceSeedAvoidsGeocoding(t *testing.T) {
	stub := &stubGeocoder{coord: Coord{Lat: 9, Lon: 9}}
	svc := NewService(stub)
	svc.Seed(map[string]Coord{"paris-france": {Lat: 48.85, Lon: 2.35}})

	coord, ok := svc.Locate(context.Background(), "Paris-France")
	if !ok || coord.Lat != 48.85 {
		t.Fatalf("seeded Locate = %+v ok=%v", coord, ok)
	}
	if got := atomic.LoadInt32(&stub.calls); got != 0 {
		t.Fatalf("geocoder called %d times, want 0 for a seeded location", got)
	}
}

func TestServiceRemembersFailures(t *testing.T) {
	stub := &stubGeocoder{err: errors.New("boom")}
	svc := NewService(stub)
	svc.minDelay = 0

	if _, ok := svc.Locate(context.Background(), "nowhere-xx"); ok {
		t.Fatal("expected ok=false on geocode failure")
	}
	if _, ok := svc.Locate(context.Background(), "nowhere-xx"); ok {
		t.Fatal("expected ok=false on a remembered failure")
	}
	if got := atomic.LoadInt32(&stub.calls); got != 1 {
		t.Fatalf("geocoder called %d times, want 1 (failure remembered)", got)
	}
}

func TestServiceSeedJSON(t *testing.T) {
	svc := NewService(&stubGeocoder{})
	if err := svc.SeedJSON([]byte(`{"berlin-germany":{"lat":52.52,"lon":13.40}}`)); err != nil {
		t.Fatalf("SeedJSON error: %v", err)
	}
	if coord, ok := svc.Locate(context.Background(), "berlin-germany"); !ok || coord.Lat != 52.52 {
		t.Fatalf("seeded JSON Locate = %+v ok=%v", coord, ok)
	}
	// Empty input is a no-op, not an error.
	if err := svc.SeedJSON([]byte("   ")); err != nil {
		t.Fatalf("SeedJSON empty error: %v", err)
	}
}

func TestEmbeddedSeedLoads(t *testing.T) {
	svc := NewService(&stubGeocoder{})
	if err := svc.SeedJSON(SeedData()); err != nil {
		t.Fatalf("embedded seed failed to load: %v", err)
	}
	// A well-known location from the committed seed should resolve with no call.
	if _, ok := svc.Locate(context.Background(), "london-uk"); !ok {
		t.Fatal("expected london-uk to be in the embedded seed")
	}
}
