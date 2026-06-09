package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"groupie-tracker/internal/geo"
	"groupie-tracker/internal/groupie"
	"groupie-tracker/internal/web"
)

// shutdownTimeout bounds how long graceful shutdown waits for in-flight
// requests before forcing the server to stop.
const shutdownTimeout = 10 * time.Second

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Stop on Ctrl+C or SIGTERM so the server and worker shut down cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		catalog  web.Catalog
		searcher web.Searcher
		worker   *groupie.SearchWorker
	)
	if loaded, err := loadCatalog(ctx); err != nil {
		log.Printf("could not load Groupie Tracker API data: %v", err)
	} else {
		catalog = loaded
		worker = groupie.NewSearchWorker(loaded)
		searcher = worker
	}

	// The geolocalization service geocodes concert locations on demand, warmed
	// by the committed seed cache so known places resolve without a network call.
	locator := geo.NewService(geo.NewNominatimGeocoder())
	if err := locator.SeedJSON(geo.SeedData()); err != nil {
		log.Printf("could not load geocode seed cache: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           routes(catalog, searcher, locator),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Run the listener in the background so main can wait on the signal.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("groupie-tracker listening on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Fatal(err)
	case <-ctx.Done():
		log.Print("shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	if worker != nil {
		worker.Close()
	}
}

func loadCatalog(ctx context.Context) (*groupie.Catalog, error) {
	client := groupie.NewClient("", groupie.DefaultClientTimeout)
	data, err := client.FetchAll(ctx)
	if err != nil {
		return nil, err
	}
	return groupie.NewCatalog(data)
}

func routes(catalog web.Catalog, searcher web.Searcher, locator web.Locator) http.Handler {
	return web.New(catalog, searcher, locator).Handler()
}
