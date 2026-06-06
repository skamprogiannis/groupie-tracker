package groupie

import (
	"sort"
	"strconv"
	"strings"
)

// Search returns the catalog entries that match the free-text query and pass
// every filter in q, ordered by q.Sort. An empty SearchQuery returns the whole
// catalog in default (name) order.
func (c *Catalog) Search(q SearchQuery) []SearchResult {
	if c == nil {
		return nil
	}

	text := strings.ToLower(strings.TrimSpace(q.Text))
	results := make([]SearchResult, 0, len(c.artists))
	for _, artist := range c.artists {
		detail := ""
		if text != "" {
			matched, matchDetail := artistMatchesWithDetail(artist, text)
			if !matched {
				continue
			}
			detail = matchDetail
		}
		if !passesFilters(artist, q) {
			continue
		}

		result := summaryToResult(summarize(artist))
		result.MatchedDetail = detail
		results = append(results, result)
	}

	sortResults(results, q.Sort)
	return results
}

func summaryToResult(s ArtistSummary) SearchResult {
	return SearchResult{
		ID:            s.ID,
		Name:          s.Name,
		Image:         s.Image,
		CreationDate:  s.CreationDate,
		FirstAlbum:    s.FirstAlbum,
		MemberSummary: s.MemberSummary,
		MemberCount:   s.MemberCount,
		LocationCount: s.LocationCount,
	}
}

func passesFilters(artist ArtistDetail, q SearchQuery) bool {
	if q.MinYear > 0 && artist.CreationDate < q.MinYear {
		return false
	}
	if q.MaxYear > 0 && artist.CreationDate > q.MaxYear {
		return false
	}

	members := len(artist.Members)
	if q.MinMembers > 0 && members < q.MinMembers {
		return false
	}
	if q.MaxMembers > 0 && members > q.MaxMembers {
		return false
	}

	if q.Country != "" && !hasCountry(artist.Locations, q.Country) {
		return false
	}
	return true
}

func hasCountry(locations []string, country string) bool {
	country = strings.ToLower(strings.TrimSpace(country))
	for _, location := range locations {
		if strings.ToLower(countryOf(location)) == country {
			return true
		}
	}
	return false
}

func sortResults(results []SearchResult, order string) {
	switch order {
	case "newest":
		sort.SliceStable(results, func(i, j int) bool {
			return results[i].CreationDate > results[j].CreationDate
		})
	case "oldest":
		sort.SliceStable(results, func(i, j int) bool {
			return results[i].CreationDate < results[j].CreationDate
		})
	case "members":
		sort.SliceStable(results, func(i, j int) bool {
			return results[i].MemberCount > results[j].MemberCount
		})
	default: // "name"
		sort.SliceStable(results, func(i, j int) bool {
			return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
		})
	}
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

// FormatLocation turns a raw "city-country" slug into a human label for
// display, e.g. "los_angeles-usa" -> "Los Angeles, USA".
func FormatLocation(raw string) string {
	return formatLocationDetail(raw)
}

// formatLocationDetail turns a "city-country" slug into a human label, e.g.
// "los_angeles-usa" -> "Los Angeles, USA".
func formatLocationDetail(raw string) string {
	if raw == "" {
		return raw
	}

	city, country, hasCountry := strings.Cut(raw, "-")
	if !hasCountry {
		return titleizeSlug(raw)
	}
	return titleizeSlug(city) + ", " + formatCountry(country)
}

// formatCountry uppercases short country codes (usa, uk, uae) and title-cases
// the rest (finland -> Finland).
func formatCountry(slug string) string {
	spaced := strings.ReplaceAll(slug, "_", " ")
	if len(strings.ReplaceAll(spaced, " ", "")) <= 3 {
		return strings.ToUpper(spaced)
	}
	return titleizeSlug(slug)
}

func titleizeSlug(slug string) string {
	cleaned := strings.ReplaceAll(slug, "_", " ")
	words := strings.Fields(cleaned)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

func containsFolded(value string, query string) bool {
	return strings.Contains(strings.ToLower(value), query)
}
