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
		if normalizedQuery == "" {
			results = append(results, SearchResult{
				ID:           artist.ID,
				Name:         artist.Name,
				Image:        artist.Image,
				CreationDate: artist.CreationDate,
				FirstAlbum:   artist.FirstAlbum,
			})
		} else if matched, detail := artistMatchesWithDetail(artist, normalizedQuery); matched {
			results = append(results, SearchResult{
				ID:            artist.ID,
				Name:          artist.Name,
				Image:         artist.Image,
				CreationDate:  artist.CreationDate,
				FirstAlbum:    artist.FirstAlbum,
				MatchedDetail: detail,
			})
		}
	}
	return results
}

func artistMatchesWithDetail(artist ArtistDetail, query string) (bool, string) {
	// Check name, album, creation date
	if containsFolded(artist.Name, query) {
		return true, ""
	}
	if containsFolded(artist.FirstAlbum, query) {
		return true, artist.FirstAlbum
	}
	if containsFolded(strconv.Itoa(artist.CreationDate), query) {
		return true, strconv.Itoa(artist.CreationDate)
	}

	// Check members
	for _, member := range artist.Members {
		if containsFolded(member, query) {
			return true, member
		}
	}

	// Check locations
	for _, location := range artist.Locations {
		if containsFolded(location, query) {
			return true, formatLocationDetail(location)
		}
	}

	// Check dates
	for _, date := range artist.Dates {
		if containsFolded(date, query) {
			return true, date
		}
	}

	// Check relations (location - dates mapping)
	for location, dates := range artist.DatesLocations {
		if containsFolded(location, query) {
			return true, formatLocationDetail(location)
		}
		for _, date := range dates {
			if containsFolded(date, query) {
				// Return location and date together for context
				return true, formatLocationDetail(location) + " - " + date
			}
		}
	}

	return false, ""
}

func formatLocationDetail(raw string) string {
	if raw == "" {
		return raw
	}
	cleaned := strings.ReplaceAll(strings.ReplaceAll(raw, "_", " "), "-", " ")
	words := strings.Fields(cleaned)
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
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
