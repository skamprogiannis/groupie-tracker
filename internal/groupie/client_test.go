package groupie

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientFetchAllSuccess(t *testing.T) {
	server := httptest.NewServer(groupieAPIHandler(nil))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, time.Second)
	data, err := client.FetchAll(context.Background())
	if err != nil {
		t.Fatalf("FetchAll returned error: %v", err)
	}

	if data.Links.Artists != server.URL+"/artists" {
		t.Fatalf("links artists = %q, want %q", data.Links.Artists, server.URL+"/artists")
	}
	if len(data.Artists) != 1 || data.Artists[0].Name != "Queen" {
		t.Fatalf("artists = %#v, want Queen fixture", data.Artists)
	}
	if len(data.Locations) != 1 || data.Locations[0].Locations[0] != "london-uk" {
		t.Fatalf("locations = %#v, want london-uk fixture", data.Locations)
	}
	if len(data.Dates) != 1 || data.Dates[0].Dates[0] != "14-12-1973" {
		t.Fatalf("dates = %#v, want first fixture date", data.Dates)
	}
	if len(data.Relations) != 1 || data.Relations[0].DatesLocations["london-uk"][0] != "14-12-1973" {
		t.Fatalf("relations = %#v, want london relation fixture", data.Relations)
	}
}

func TestClientFetchAllReturnsNon2xxErrors(t *testing.T) {
	server := httptest.NewServer(groupieAPIHandler(map[string]http.HandlerFunc{
		"/artists": func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "upstream failed", http.StatusBadGateway)
		},
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, time.Second)
	_, err := client.FetchAll(context.Background())
	if err == nil {
		t.Fatal("FetchAll returned nil error, want non-2xx error")
	}
	if !strings.Contains(err.Error(), "unexpected status 502") {
		t.Fatalf("error = %q, want status detail", err.Error())
	}
}

func TestClientFetchAllReturnsDecodeErrors(t *testing.T) {
	server := httptest.NewServer(groupieAPIHandler(map[string]http.HandlerFunc{
		"/locations": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"index":`))
		},
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, time.Second)
	_, err := client.FetchAll(context.Background())
	if err == nil {
		t.Fatal("FetchAll returned nil error, want decode error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("error = %q, want decode detail", err.Error())
	}
}

func TestClientFetchAllReturnsTimeoutErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, time.Millisecond)
	_, err := client.FetchAll(context.Background())
	if err == nil {
		t.Fatal("FetchAll returned nil error, want timeout error")
	}
	if !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("error = %q, want fetch detail", err.Error())
	}
}

func groupieAPIHandler(overrides map[string]http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if overrides != nil {
			if handler, ok := overrides[r.URL.Path]; ok {
				handler(w, r)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`{"artists":"` + serverURL(r) + `/artists","locations":"` + serverURL(r) + `/locations","dates":"` + serverURL(r) + `/dates","relation":"` + serverURL(r) + `/relation"}`))
		case "/artists":
			_, _ = w.Write([]byte(`[{"id":1,"image":"queen.jpeg","name":"Queen","members":["Freddie Mercury","Brian May"],"creationDate":1970,"firstAlbum":"14-12-1973","locations":"` + serverURL(r) + `/locations/1","concertDates":"` + serverURL(r) + `/dates/1","relations":"` + serverURL(r) + `/relation/1"}]`))
		case "/locations":
			_, _ = w.Write([]byte(`{"index":[{"id":1,"locations":["london-uk"],"dates":"` + serverURL(r) + `/dates/1"}]}`))
		case "/dates":
			_, _ = w.Write([]byte(`{"index":[{"id":1,"dates":["14-12-1973"]}]}`))
		case "/relation":
			_, _ = w.Write([]byte(`{"index":[{"id":1,"datesLocations":{"london-uk":["14-12-1973"]}}]}`))
		default:
			http.NotFound(w, r)
		}
	})
}

func serverURL(r *http.Request) string {
	return "http://" + r.Host
}
