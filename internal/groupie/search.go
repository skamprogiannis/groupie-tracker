package groupie

import (
	"strconv"
	"strings"
)

func (c *Catalog) Search(query string) []SearchResult {
	if c == nil {
		return nil
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	results := make([]SearchResult, 0, len(c.artists))
	for _, artist := range c.artists {
		if normalizedQuery == "" || artistMatches(artist, normalizedQuery) {
			results = append(results, SearchResult{
				ID:           artist.ID,
				Name:         artist.Name,
				Image:        artist.Image,
				CreationDate: artist.CreationDate,
				FirstAlbum:   artist.FirstAlbum,
			})
		}
	}
	return results
}

func artistMatches(artist ArtistDetail, query string) bool {
	if containsFolded(artist.Name, query) ||
		containsFolded(artist.FirstAlbum, query) ||
		containsFolded(strconv.Itoa(artist.CreationDate), query) {
		return true
	}

	return containsAnyFolded(artist.Members, query) ||
		containsAnyFolded(artist.Locations, query) ||
		containsAnyFolded(artist.Dates, query) ||
		relationMatches(artist.DatesLocations, query)
}

func relationMatches(datesLocations map[string][]string, query string) bool {
	for location, dates := range datesLocations {
		if containsFolded(location, query) || containsAnyFolded(dates, query) {
			return true
		}
	}
	return false
}

func containsAnyFolded(values []string, query string) bool {
	for _, value := range values {
		if containsFolded(value, query) {
			return true
		}
	}
	return false
}

func containsFolded(value string, query string) bool {
	return strings.Contains(strings.ToLower(value), query)
}
