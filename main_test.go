package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"groupie-tracker/internal/groupie"
)

func TestRoutesHome(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	routes(mainTestCatalog{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Queen") {
		t.Fatalf("body = %q, want Queen catalog entry", rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	routes(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

type mainTestCatalog struct{}

func (mainTestCatalog) Artists() []groupie.ArtistSummary {
	return []groupie.ArtistSummary{
		{ID: 1, Name: "Queen", Image: "queen.jpeg", CreationDate: 1970, FirstAlbum: "14-12-1973"},
	}
}

func (mainTestCatalog) ArtistByID(id int) (groupie.ArtistDetail, bool) {
	return groupie.ArtistDetail{}, false
}

func (mainTestCatalog) Search(query string) []groupie.SearchResult {
	return nil
}
