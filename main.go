package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"groupie-tracker/internal/groupie"
	"groupie-tracker/internal/web"
)

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var catalog web.Catalog
	loadedCatalog, err := loadCatalog(ctx)
	if err != nil {
		log.Printf("could not load Groupie Tracker API data: %v", err)
	} else {
		catalog = loadedCatalog
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           routes(catalog),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("groupie-tracker listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
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

func routes(catalog web.Catalog) http.Handler {
	return web.New(catalog).Handler()
}
